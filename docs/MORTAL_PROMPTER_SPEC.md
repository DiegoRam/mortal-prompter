# 🎮 MORTAL PROMPTER - Especificación Completa

> **Documento de especificación para crear con Claude Code**
> 
> Ejecutar: `claude -p "$(cat MORTAL_PROMPTER_SPEC.md)"`

---

## Resumen del Proyecto

Crea una aplicación CLI en **Go** llamada `mortal-prompter` que orquesta un loop de desarrollo y code review entre **Claude Code** y **OpenAI Codex**.

### Concepto

La herramienta actúa como un "árbitro" entre dos LLMs en un combate estilo Mortal Kombat:

- **CLAUDE CODE** (Fighter 1): Ejecuta tareas de desarrollo/implementación
- **CODEX** (Fighter 2): Revisa el código y encuentra issues

El loop continúa hasta que Codex no encuentre más problemas o se alcance el límite de iteraciones.

### Flujo Principal

```
Usuario envía prompt inicial
        ↓
┌───────────────────────┐
│   ROUND N             │
├───────────────────────┤
│ 1. Claude Code ejecuta│
│ 2. Captura git diff   │
│ 3. Codex revisa diff  │
│ 4. ¿Hay issues?       │
│    - Sí → nuevo round │
│    - No → FINISH HIM! │
└───────────────────────┘
        ↓
Commit final + reporte
```

---

## Estructura del Proyecto

```
mortal-prompter/
├── cmd/
│   └── mortal-prompter/
│       └── main.go               # Entry point
├── internal/
│   ├── orchestrator/
│   │   └── orchestrator.go       # Loop principal de "combate"
│   ├── fighters/
│   │   ├── claude.go             # Wrapper para claude CLI
│   │   └── codex.go              # Wrapper para codex CLI
│   ├── git/
│   │   └── git.go                # Operaciones git (diff, commit, etc)
│   ├── logger/
│   │   └── logger.go             # Logging a terminal y archivo
│   └── config/
│       └── config.go             # Configuración y flags
├── pkg/
│   └── types/
│       └── types.go              # Tipos compartidos
├── scripts/
│   └── install.sh                # Script de instalación universal (Fase 2)
├── .github/
│   └── workflows/
│       └── release.yml           # GitHub Actions para releases (Fase 2)
├── Makefile                      # Build tasks
├── .goreleaser.yml               # Configuración GoReleaser (Fase 2)
├── go.mod
├── go.sum
├── README.md
├── LICENSE                       # MIT License
└── .gitignore
```

---

## Especificaciones Técnicas

### 1. Entry Point (cmd/mortal-prompter/main.go)

```go
// Variables de versión (para builds)
var (
    Version   = "dev"
    BuildTime = "unknown"
)

// Flags requeridos:
// -p, --prompt string      Prompt inicial para Claude Code (requerido)
// -d, --dir string         Directorio de trabajo (default: ".")
// -m, --max-iterations int Máximo de iteraciones (default: 10)
// -i, --interactive        Modo interactivo, pide confirmación cada iteración
// -v, --verbose            Output detallado
// -o, --output string      Directorio para logs y reportes (default: ".mortal-prompter")
// --auto-commit            Hace commit automático cuando termina exitosamente
// --commit-message string  Mensaje base para commits (default: "feat: implemented via mortal-prompter")
// --version                Muestra versión y sale

// Ejemplo de uso:
// mortal-prompter -p "implementa autenticación JWT" --auto-commit -v
// mortal-prompter --prompt "agrega tests unitarios para el módulo users" -m 5 -i
```

### 2. Orchestrator (internal/orchestrator/orchestrator.go)

El orquestador maneja el "combate" entre los dos LLMs:

```go
type Orchestrator struct {
    claude      *fighters.Claude
    codex       *fighters.Codex
    git         *git.Git
    logger      *logger.Logger
    config      *config.Config
    rounds      []Round
}

type Round struct {
    Number          int
    ClaudePrompt    string
    ClaudeOutput    string
    GitDiff         string
    CodexReview     string
    HasIssues       bool
    Issues          []string
    Duration        time.Duration
    Timestamp       time.Time
}

// Métodos principales:
// - Start(initialPrompt string) error
// - runRound(prompt string) (*Round, error)
// - shouldContinue(round *Round) bool
// - requestManualConfirmation() bool
// - generateFinalReport() error
```

**Lógica del loop:**

1. Ejecutar Claude Code con el prompt
2. Esperar a que termine (capturar stdout/stderr)
3. Obtener `git diff` de cambios unstaged
4. Si no hay cambios, loggear warning y continuar/terminar
5. Enviar diff a Codex con prompt de review
6. Parsear respuesta de Codex para detectar issues
7. Si hay issues, construir nuevo prompt para Claude con los issues
8. Si no hay issues, terminar exitosamente
9. Si iteración >= 10, pedir confirmación manual
10. Repetir

### 3. Claude Fighter (internal/fighters/claude.go)

```go
type Claude struct {
    workDir string
    logger  *logger.Logger
}

// Ejecuta claude code CLI
// Comando: claude -p "<prompt>" --dangerously-skip-permissions
// 
// El flag --dangerously-skip-permissions es necesario para ejecución no interactiva
// 
// Métodos:
// - Execute(prompt string) (output string, err error)
// - buildPrompt(basePrompt string, previousIssues []string) string
```

**Prompt template para Claude (cuando hay issues previos):**

```
CONTEXTO: Estás en una sesión de code review iterativo.

ISSUES ENCONTRADOS EN LA REVISIÓN ANTERIOR:
{{range .Issues}}
- {{.}}
{{end}}

TAREA: Corrige los issues mencionados arriba. 
No expliques los cambios, solo implementa las correcciones.
```

### 4. Codex Fighter (internal/fighters/codex.go)

```go
type Codex struct {
    workDir string
    logger  *logger.Logger
}

// Ejecuta codex CLI
// Comando: codex -p "<prompt>"
//
// Métodos:
// - Review(gitDiff string) (review *ReviewResult, err error)
// - parseReviewOutput(output string) *ReviewResult

type ReviewResult struct {
    HasIssues   bool
    Issues      []string
    RawOutput   string
}
```

**Prompt template para Codex review:**

```
Actúa como un code reviewer senior extremadamente exigente.

Revisa el siguiente diff de git y encuentra TODOS los problemas:
- Bugs o errores lógicos
- Vulnerabilidades de seguridad
- Malas prácticas
- Código duplicado
- Falta de manejo de errores
- Problemas de performance
- Violaciones de convenciones de Go

GIT DIFF:
```diff
{{.GitDiff}}
```

INSTRUCCIONES DE RESPUESTA:
- Si NO encuentras issues, responde EXACTAMENTE: "LGTM: No issues found"
- Si encuentras issues, lista cada uno en formato:
  ISSUE: [descripción del problema]
  
Sé conciso y específico. No incluyas sugerencias opcionales, solo problemas reales.
```

### 5. Git Utils (internal/git/git.go)

```go
type Git struct {
    workDir string
    logger  *logger.Logger
}

// Métodos:
// - GetUnstagedDiff() (string, error)         // git diff
// - GetStagedDiff() (string, error)           // git diff --staged
// - StageAll() error                          // git add -A
// - Commit(message string) error              // git commit -m "..."
// - GetCurrentBranch() (string, error)
// - HasUncommittedChanges() (bool, error)
```

### 6. Logger (internal/logger/logger.go)

```go
type Logger struct {
    verbose     bool
    logFile     *os.File
    outputDir   string
}

// Métodos:
// - RoundStart(number int)
// - FighterEnter(name string)
// - FighterAction(action string)
// - FighterFinish(name string, duration time.Duration)
// - IssuesFound(issues []string)
// - NoIssues()
// - FinalVictory(totalRounds int, totalDuration time.Duration)
// - Error(err error)
// - Info(msg string)
// - Debug(msg string)  // solo si verbose
```

**Output a terminal con colores estilo arcade/fighting game:**

```
═══════════════════════════════════════════════════════════
🎮 MORTAL PROMPTER - ROUND 1
═══════════════════════════════════════════════════════════

🥊 CLAUDE CODE enters the arena...
⏳ Executing task...
✅ CLAUDE CODE finishes! (took 45s)

📝 Changes detected: 5 files modified

🥊 CODEX enters the arena...
🔍 Reviewing changes...
⚠️  CODEX found 3 issues!

   ISSUE 1: Missing error handling in auth.go:45
   ISSUE 2: SQL injection vulnerability in users.go:23
   ISSUE 3: Unused variable in main.go:12

🔄 Preparing next round...
═══════════════════════════════════════════════════════════
```

### 7. Config (internal/config/config.go)

```go
type Config struct {
    Prompt           string
    WorkDir          string
    MaxIterations    int
    Interactive      bool
    Verbose          bool
    OutputDir        string
    AutoCommit       bool
    CommitMessage    string
}

// ParseFlags() *Config - parsea flags de CLI
// Validate() error - valida configuración
```

---

## Output Files

### Log File (.mortal-prompter/session-{timestamp}.log)

```
[2024-01-15 10:30:00] SESSION START
[2024-01-15 10:30:00] Initial prompt: "implementa autenticación JWT"
[2024-01-15 10:30:00] Working directory: /home/user/project
[2024-01-15 10:30:00] Max iterations: 10

[2024-01-15 10:30:01] ROUND 1 START
[2024-01-15 10:30:01] Executing Claude Code...
[2024-01-15 10:30:45] Claude Code finished (44s)
[2024-01-15 10:30:45] Git diff: 5 files changed, 234 insertions, 12 deletions
[2024-01-15 10:30:46] Executing Codex review...
[2024-01-15 10:31:02] Codex review finished (16s)
[2024-01-15 10:31:02] Issues found: 3
[2024-01-15 10:31:02] ROUND 1 END

... más rounds ...

[2024-01-15 10:35:00] SESSION END - SUCCESS
[2024-01-15 10:35:00] Total rounds: 3
[2024-01-15 10:35:00] Total duration: 5m00s
```

### Final Report (.mortal-prompter/report-{timestamp}.md)

```markdown
# 🎮 Mortal Prompter - Battle Report

## Summary
- **Initial Prompt:** implementa autenticación JWT
- **Total Rounds:** 3
- **Total Duration:** 5m 00s
- **Result:** ✅ SUCCESS - FLAWLESS VICTORY

## Round History

### Round 1
**Claude Code Task:** implementa autenticación JWT
**Duration:** 44s
**Files Changed:** 5

**Codex Review:** ⚠️ 3 issues found
1. Missing error handling in auth.go:45
2. SQL injection vulnerability in users.go:23
3. Unused variable in main.go:12

---

### Round 2
**Claude Code Task:** Fix issues from previous review
**Duration:** 30s
**Files Changed:** 3

**Codex Review:** ⚠️ 1 issue found
1. JWT expiration not configurable

---

### Round 3
**Claude Code Task:** Fix issues from previous review
**Duration:** 15s
**Files Changed:** 1

**Codex Review:** ✅ LGTM - No issues found

---

## Final Changes
(git diff final incluido aquí)

## Files Modified
- internal/auth/auth.go
- internal/users/users.go
- cmd/server/main.go
- config/config.go
- go.mod
```

---

## Manejo de Errores

| Escenario | Acción |
|-----------|--------|
| Claude Code falla | Loggear error, mostrar output, preguntar si reintentar |
| Codex falla | Loggear error, mostrar output, preguntar si reintentar |
| Git falla | Abortar con mensaje claro |
| Timeout (5 min por fighter) | Abortar round, preguntar si continuar |
| Sin cambios después de Claude | Warning, preguntar si continuar |

---

## Confirmación Manual (después de 10 iteraciones)

```
═══════════════════════════════════════════════════════════
⚠️  MAXIMUM ITERATIONS REACHED (10)
═══════════════════════════════════════════════════════════

The battle has gone on for 10 rounds without resolution.
Current issues still pending:
  1. [issue 1]
  2. [issue 2]

Options:
  [c] Continue for 5 more rounds
  [s] Stop and commit current state
  [a] Abort without committing

Your choice: _
```

---

## Dependencias Go

```go
// go.mod
module github.com/diegoram/mortal-prompter

go 1.21

require (
    github.com/spf13/cobra v1.8.0          // CLI framework
    github.com/fatih/color v1.16.0         // Terminal colors
    github.com/briandowns/spinner v1.23.0  // Loading spinners
)
```

---

## Notas de Implementación

1. **Ejecución de CLIs externos:** Usar `os/exec` con timeout context
2. **Captura de output:** Capturar tanto stdout como stderr
3. **Parsing de issues:** Buscar líneas que empiecen con "ISSUE:" en output de Codex
4. **Detección de LGTM:** Buscar "LGTM" o "No issues" en output de Codex
5. **Colores en terminal:** Usar fatih/color, detectar si terminal soporta colores
6. **Spinners:** Mostrar spinner mientras los fighters trabajan
7. **Signals:** Manejar SIGINT/SIGTERM para cleanup graceful

---

## Makefile

```makefile
APP_NAME := mortal-prompter
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

.PHONY: build install clean test release-dry build-all

# Build para la plataforma actual
build:
	go build $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/mortal-prompter

# Instalar en el sistema
install:
	go install $(LDFLAGS) ./cmd/mortal-prompter

# Build para todas las plataformas
build-all:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-darwin-amd64 ./cmd/mortal-prompter
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(APP_NAME)-darwin-arm64 ./cmd/mortal-prompter
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-linux-amd64 ./cmd/mortal-prompter
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(APP_NAME)-linux-arm64 ./cmd/mortal-prompter
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-windows-amd64.exe ./cmd/mortal-prompter

# Limpiar builds
clean:
	rm -rf bin/
	rm -rf dist/

# Correr tests
test:
	go test -v ./...

# Dry run de release (para probar sin publicar)
release-dry:
	goreleaser release --snapshot --clean
```

---

## 📦 Distribución y Empaquetado (Fase 2 - No bloqueante)

> **NOTA:** Esta sección NO es requerida para la implementación inicial.
> Implementar solo después de que la funcionalidad core esté completa y testeada.

### GoReleaser (.goreleaser.yml)

```yaml
project_name: mortal-prompter

before:
  hooks:
    - go mod tidy

builds:
  - id: mortal-prompter
    main: ./cmd/mortal-prompter
    binary: mortal-prompter
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.Version={{.Version}}
      - -X main.BuildTime={{.Date}}

archives:
  - id: default
    format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip
    files:
      - README.md
      - LICENSE

checksum:
  name_template: 'checksums.txt'

snapshot:
  name_template: "{{ .Tag }}-next"

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^ci:'

brews:
  - name: mortal-prompter
    repository:
      owner: diegoram
      name: homebrew-tap
    homepage: "https://github.com/diegoram/mortal-prompter"
    description: "CLI que orquesta code review entre Claude Code y Codex"
    license: "MIT"
    install: |
      bin.install "mortal-prompter"
    test: |
      system "#{bin}/mortal-prompter", "--version"

release:
  github:
    owner: diegoram
    name: mortal-prompter
  draft: false
  prerelease: auto
```

### GitHub Actions (.github/workflows/release.yml)

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Run tests
        run: go test -v ./...

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v5
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          HOMEBREW_TAP_GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}
```

### Script de instalación (scripts/install.sh)

```bash
#!/bin/bash
set -e

# Colores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}"
echo "═══════════════════════════════════════════════════════════"
echo "🎮 MORTAL PROMPTER - Installer"
echo "═══════════════════════════════════════════════════════════"
echo -e "${NC}"

# Detectar OS y arquitectura
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo -e "${RED}Arquitectura no soportada: $ARCH${NC}"; exit 1 ;;
esac

case $OS in
    darwin|linux) ;;
    *) echo -e "${RED}OS no soportado: $OS${NC}"; exit 1 ;;
esac

# Obtener última versión
REPO="diegoram/mortal-prompter"
LATEST_VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_VERSION" ]; then
    echo -e "${RED}No se pudo obtener la última versión${NC}"
    exit 1
fi

echo -e "${YELLOW}Instalando mortal-prompter $LATEST_VERSION para $OS/$ARCH...${NC}"

# Construir URL de descarga
FILENAME="mortal-prompter_${LATEST_VERSION#v}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_VERSION/$FILENAME"

# Directorio temporal
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# Descargar y extraer
echo "Descargando desde $DOWNLOAD_URL..."
curl -sL "$DOWNLOAD_URL" -o "$TMP_DIR/mortal-prompter.tar.gz"
tar -xzf "$TMP_DIR/mortal-prompter.tar.gz" -C "$TMP_DIR"

# Instalar
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
    echo -e "${YELLOW}Necesita permisos de administrador para instalar en $INSTALL_DIR${NC}"
    sudo mv "$TMP_DIR/mortal-prompter" "$INSTALL_DIR/"
else
    mv "$TMP_DIR/mortal-prompter" "$INSTALL_DIR/"
fi

chmod +x "$INSTALL_DIR/mortal-prompter"

echo -e "${GREEN}"
echo "═══════════════════════════════════════════════════════════"
echo "✅ INSTALLATION COMPLETE!"
echo "═══════════════════════════════════════════════════════════"
echo -e "${NC}"
echo "Ejecuta 'mortal-prompter --help' para comenzar"
echo ""
echo -e "${YELLOW}FIGHT!${NC} 🥊"
```

### Métodos de instalación para colaboradores (Fase 2)

```bash
# Opción 1: Homebrew (macOS/Linux)
brew tap diegoram/tap
brew install mortal-prompter

# Opción 2: Script directo (macOS/Linux)
curl -sSL https://raw.githubusercontent.com/diegoram/mortal-prompter/main/scripts/install.sh | bash

# Opción 3: Go install (requiere Go)
go install github.com/diegoram/mortal-prompter/cmd/mortal-prompter@latest

# Opción 4: Descarga manual desde GitHub Releases
```

---

## Ejemplos de Uso

```bash
# Básico
mortal-prompter -p "implementa endpoint REST para usuarios"

# Con auto-commit
mortal-prompter -p "agrega validación de inputs" --auto-commit

# Modo verbose e interactivo
mortal-prompter -p "refactoriza el módulo de auth" -v -i

# En directorio específico
mortal-prompter -p "agrega tests" -d ./backend --max-iterations 5

# Ver versión
mortal-prompter --version
```

---

## README.md sugerido

```markdown
# 🎮 Mortal Prompter

> *"FINISH HIM!"* - Cuando tu código finalmente pasa code review

CLI que orquesta un loop de desarrollo y code review entre **Claude Code** y **OpenAI Codex**.

## ¿Cómo funciona?

1. Enviás un prompt de desarrollo
2. Claude Code implementa los cambios
3. Codex revisa el código y encuentra issues
4. Claude Code corrige los issues
5. Repeat hasta **FLAWLESS VICTORY** 🏆

## Instalación

```bash
# Con Go
go install github.com/diegoram/mortal-prompter/cmd/mortal-prompter@latest

# Con Homebrew
brew tap diegoram/tap && brew install mortal-prompter

# Script directo
curl -sSL https://raw.githubusercontent.com/diegoram/mortal-prompter/main/scripts/install.sh | bash
```

## Uso

```bash
mortal-prompter -p "tu prompt aquí"
```

## Requisitos

- Git instalado y configurado
- Claude Code CLI instalado (`claude`)
- OpenAI Codex CLI instalado (`codex`)

## Licencia

MIT
```

---

## Comando de Inicialización

Una vez descargado este archivo, ejecuta:

```bash
# Crear directorio del proyecto
mkdir mortal-prompter && cd mortal-prompter

# Inicializar con Claude Code
claude -p "Lee el archivo MORTAL_PROMPTER_SPEC.md y crea el proyecto completo siguiendo todas las especificaciones. Crea primero la estructura de directorios, luego implementa cada archivo en orden. Comienza por go.mod, luego los tipos, después la configuración, y finalmente los componentes principales. La Fase 2 (distribución) déjala para después."
```

---

**FIGHT!** 🥊🎮
