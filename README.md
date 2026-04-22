# Claude Code

Configuracion completa de Claude Code: agentes, skills, hooks, settings. Todo lo necesario para restaurar el entorno de desarrollo AI en cualquier sistema.

## Estructura

```
claude-code/
├── agents/                      # Definiciones de agentes (system prompts)
│   ├── orchestrator.md          # Hub central — coordina todo el equipo
│   ├── architect.md             # Diseño, investigación, planning, auditoría
│   ├── implementer.md           # Código de producción
│   ├── tester.md                # Tests con factory mocks
│   ├── qa.md                    # Validación contra AC / definición de AC
│   ├── security.md              # Auditoría de seguridad (OWASP, CWE, ASVS)
│   ├── delivery.md              # Branch, changelog, version, commit, PR
│   ├── reviewer.md              # Review de PRs en GitHub
│   ├── init.md                  # Bootstrap de CLAUDE.md en cualquier repo
│   ├── diagrammer.md            # Diagramas Excalidraw con render-validate loop
│   └── agent-builder.md         # Crea y mejora agentes y skills
│
├── skills/                      # Entry points (slash commands)
│   ├── issue.md ... tmux.md     # 20 skills individuales
│   └── excalidraw-diagram/      # Skill compleja con templates y renderer
│
├── hooks/                       # Notification scripts (one per platform)
│   ├── notify-windows.sh        # Windows → PowerShell toast notification
│   ├── notify-linux.sh          # Linux → notify-send (libnotify)
│   ├── notify-openclaw.sh       # OpenClaw → Telegram (optional, always notifies)
│   └── notify-openclaw-progress.sh  # OpenClaw → Telegram (only active pipeline)
│
├── settings.json                # Respaldo de ~/.claude/settings.json
├── hooks-config.json            # Referencia de hooks configurados
├── diagram.excalidraw           # Diagrama visual del sistema completo
└── diagram_preview.png          # Preview del diagrama
```

### Destinos activos

| Source (repo) | Destino activo | Qué contiene |
|---|---|---|
| `agents/*.md` | `~/.claude/agents/` | System prompts de agentes |
| `skills/*.md` | `~/.claude/commands/` (global) | Slash commands |
| `skills/excalidraw-diagram/` | `~/.claude/skills/excalidraw-diagram/` | Skill compleja |
| `hooks/*.sh` | `~/.claude/hooks/` | Scripts de hooks |
| `settings.json` | `~/.claude/settings.json` | Settings globales |

---

## Settings

| Setting | Valor | Descripcion |
|---|---|---|
| `effortLevel` | `high` | Nivel de esfuerzo de Claude |
| `autoUpdatesChannel` | `latest` | Canal de actualizaciones |
| `skipDangerousModePermissionPrompt` | `true` | No preguntar al activar dangerous mode |
| `voiceEnabled` | `true` | Voice input habilitado |

### Permissions (allowlist)

Comandos pre-aprobados que no requieren confirmacion del usuario:
- `npx jest:*` — tests
- `netstat -ano` — diagnostico de red
- `WebFetch(github.com, pixijs.com)` — fetch de docs

---

## Hooks (Notifications)

One script per platform. Pick the one matching your OS from `hooks-config.json` and copy its `hooks` section to `~/.claude/settings.json`.

### Windows (`notify-windows.sh`)

PowerShell toast notification. No external dependencies.

```
Claude Code (hook event) → notify-windows.sh → PowerShell WinRT toast
```

**Requires:** `powershell.exe` (included in Windows). Uses AUMID `{1AC14E77-...}\powershell.exe` for toast delivery.

### Linux (`notify-linux.sh`)

Native desktop notification via libnotify.

```
Claude Code (hook event) → notify-linux.sh → notify-send
```

**Requires:** `notify-send` (package `libnotify-bin` on Debian/Ubuntu).

### Events (both platforms)

| Hook | When it fires |
|---|---|
| **Stop** | Claude finished working |
| **Notification** | Claude is waiting for input (idle, permission prompt) |
| **PostToolUseFailure** | A tool failed (intervene before it spirals) |

### OpenClaw — Telegram (optional)

Remote monitoring via OpenClaw gateway → Telegram. Requires `OPENCLAW_GATEWAY_TOKEN` and OpenClaw running on `localhost:18789`.

```
Claude Code (hook event) → bash script → POST localhost:18789/hooks/wake → Telegram
```

**Scripts:**

- **`notify-openclaw.sh`** — Global hook. Always fires. Used for Notification and PostToolUseFailure.
- **`notify-openclaw-progress.sh`** — Pipeline hook. Checks `session-docs/*/00-state.md` for active pipeline. Silent if no pipeline running.

**OpenClaw events:**

| Hook | Matcher | Script | Scope |
|---|---|---|---|
| **Stop** | — | `progress` | Orchestrator only |
| **Notification** | `idle_prompt\|permission_prompt` | `global` | Always |
| **PostToolUse** | `Write\|Edit\|Bash` | `progress` | Orchestrator only |
| **PreToolUse** | `Bash` | `progress` | Orchestrator only |
| **PostToolUseFailure** | `*` | `global` | Always |

---

## Sistema de Agentes

11 agentes especializados coordinados por un orchestrator central. Pipeline completo: issue → diseño → implementacion → testing → delivery → PR. 20 slash commands (15 rutean al orchestrator, 5 standalone).

```
Usuario ──> /skill ──> Orchestrator ──> Agente(s) ──> session-docs ──> Resultado
```

### Agentes

| Agente | Rol | Escribe código |
|--------|-----|:--------------:|
| **orchestrator** | Hub central, coordina todo el equipo | No |
| **architect** | Diseña arquitectura, investiga, planifica, audita | No |
| **implementer** | Escribe código de producción | Sí |
| **tester** | Crea y ejecuta tests (factory mocks) | Sí (tests) |
| **qa** | Valida contra AC o define AC standalone | No |
| **security** | Audita seguridad (OWASP, CWE, ASVS) | No |
| **delivery** | Documenta, versiona, branch, commit, PR | No |
| **reviewer** | Revisa PRs en GitHub | No |
| **init** | Bootstrap de CLAUDE.md en cualquier repo | No |
| **diagrammer** | Genera diagramas Excalidraw | No |
| **agent-builder** | Crea y mejora agentes y skills | Sí |

### Skills (Entry Points)

**Rutean al orchestrator (15 skills):**

```
/issue <#N | texto>        Full pipeline o batch de issues
/plan <#N | texto>         Planning + opcionalmente ejecutar cada tarea
/design <#N | texto>       Solo arquitectura
/research <topic>          Investigación de tecnologías
/test <feature>            Solo tests
/validate <feature>        Validación contra AC
/deliver <feature>         Branch, docs, changelog, version, PR
/define-ac <#N | texto>    Definir criterios de aceptación
/security <feature>        Auditoría de seguridad
/review-pr <#N>            Review de PR
/diagram <descripción>     Diagrama Excalidraw
/init                      Bootstrap del repo
/spike <descripción>       Exploración rápida
/audit [target]            Auditoría de arquitectura
/resume [feature]          Retomar pipeline interrumpido
```

**Standalone (5 skills — NO rutean al orchestrator):**

```
/status [feature]          Estado de pipelines activos
/memory <action>           Gestionar Knowledge Graph
/lint                      Health check de agentes y skills
/tmux <action>             Sesiones tmux paralelas
/kg-viewer <action>        Viewer web del Knowledge Graph
```

---

## Pipeline Principal (`/issue`)

```
0a Intake ──> 0b Specify ──> 1 Design ──> 2 Implement ──> 3 Verify ──> 4 Delivery ──> 5 GitHub
                                               ↑              │
                                               └── fail ──────┘  (max 3 iteraciones)

                                          Phase 3 — Verify (paralelo):
                                          ├── tester
                                          ├── qa (validate)
                                          └── security* (*solo si security-sensitive)
```

### Phase 0a — Intake
Input → duplicate check → Knowledge Graph query → classify (type/complexity/security-sensitive) → bootstrap check (CLAUDE.md exists?)

### Phase 0b — Specify
Investigate codebase → build spec (User Story + AC Given/When/Then + Scope) → resolve ambiguities → update GitHub issue → write `00-task-intake.md`

### Phase 1 — Design (architect)
Lee spec → investiga codebase via context7 → produce `01-architecture.md` con decisiones, assessment, archivos a modificar

### Phase 2 — Implementation (implementer)
Lee architecture + spec → investiga docs via context7 → escribe código → build/lint loop interno → `02-implementation.md`

### Phase 3 — Verify (paralelo)
Lanza tester + qa + security* en paralelo. Si alguno falla → implementer arregla → re-verify (max 3 iteraciones). Security gate: Critical/High bloquean, Medium/Low son warnings.

### Phase 4 — Delivery
Branch + CHANGELOG + version bump + CLAUDE.md memory + commit + PR → `05-delivery.md`

### Phase 5 — GitHub Update
Comment en issue con resultados detallados → mover a "In Review"

### Knowledge Save (post-pipeline)
Extrae 1-3 insights → dedup check (busqueda semantica) → create_entities en ChromaDB

---

## Otros Flujos

| Flujo | Comando | Qué hace |
|---|---|---|
| **Planning** | `/plan` | Architect analiza → breakdown en issues con AC |
| **Research** | `/research` | Investigación de tecnologías, reporte neutral |
| **Design** | `/design` | Solo arquitectura, sin implementación |
| **Spike** | `/spike` | Exploración rápida sin tests/delivery |
| **Audit** | `/audit` | Health check arquitectónico (critical/warning/info) |
| **Security** | `/security` | Auditoría OWASP/CWE/ASVS standalone |
| **Review** | `/review-pr` | Skill hace Bash → reviewer analiza → usuario aprueba |
| **Diagram** | `/diagram` | Architect analiza → diagrammer genera Excalidraw |
| **Deliver** | `/deliver` | Delivery standalone de implementación existente |
| **Resume** | `/resume` | Retoma pipeline desde último checkpoint |

---

## Session-Docs (Shared Board)

Canal de comunicación entre agentes. Cada feature: `session-docs/{feature-name}/`.

```
00-state.md              ← orchestrator (checkpoint + hot context + recovery)
00-execution-log.md      ← todos (append start/end con timestamp)
00-task-intake.md        ← orchestrator (spec con AC)
01-architecture.md       ← architect
02-implementation.md     ← implementer
03-testing.md            ← tester
04-validation.md         ← qa
04-security.md           ← security (si security-sensitive)
05-delivery.md           ← delivery
```

Siempre en `.gitignore` (`/session-docs`).

---

## Sistema de Memoria (3 Capas)

1. **ChromaDB Knowledge Graph** (cross-project) — `~/.claude/chromadb/`. Busqueda semantica con embeddings locales. Orchestrator escribe post-pipeline, lee en Phase 0a.
2. **CLAUDE.md** (per-project) — delivery extrae knowledge. 4 secciones: Architecture Decisions, Patterns & Conventions, Known Constraints, Testing Conventions.
3. **docs/knowledge.md** (per-project) — bullets planos con decisiones y patrones.

---

## Restaurar

```bash
git clone <repo> && cd devtools-config && ./scripts/setup.sh
```

El setup despliega:
- Agentes → `~/.claude/agents/`
- Skills → `~/.claude/commands/` + `~/.claude/skills/excalidraw-diagram/`
- ChromaDB MCP → `~/.claude/chromadb-mcp/`
- MCP servers: "memory" (ChromaDB) + "context7"

Para hooks (manual):
```bash
mkdir -p ~/.claude/hooks
cp claude-code/hooks/*.sh ~/.claude/hooks/
chmod +x ~/.claude/hooks/*.sh
# Elegir backend en hooks-config.json (desktop o openclaw)
# Copiar la sección "hooks" del backend elegido a ~/.claude/settings.json
```

Verificar sincronización: `/lint`
