# Política de contenido del Knowledge Graph

**Ámbito.** Esta política rige qué puede almacenarse en el Knowledge Graph (KG) del sistema `claude-dev-team`, tanto en la KG local de cada desarrollador (`~/.claude/chromadb/`) como en cualquier export compartido con el equipo.

**Principio rector.** El KG es **memoria técnica compartible**. Todo lo que entra debe ser útil para otro desarrollador del equipo y seguro para circular entre máquinas.

---

## ✅ Permitido

- **Patrones de código** reutilizables (convenciones de un framework, soluciones recurrentes).
- **Gotchas y trampas** de librerías, runtimes o herramientas (`pnpm`, `Supabase`, `shadcn`, `Drizzle`, etc.).
- **Decisiones arquitectónicas** con rationale técnico (por qué X en lugar de Y, con restricciones técnicas observadas).
- **Inventarios técnicos** de un proyecto (listado de servicios, puertos, endpoints públicos, estructura de carpetas).
- **Comandos útiles** específicos de un stack (builds, migraciones, debugging).
- **Convenciones técnicas** de un proyecto (naming, layout, testing).

## ❌ Prohibido

- **Datos personales**: nombres de personas (incluyendo el del propio dev), roles, responsabilidades, preferencias personales.
- **Feedback personalizado**: instrucciones dirigidas a un usuario específico ("a Mario le gusta que…").
- **Credenciales y secretos**: tokens, API keys, URLs de servicios privados, IPs internas, rutas absolutas con usuario (`C:/Users/mario/...`).
- **Datos de clientes / stakeholders**: nombres de empresas cliente, contactos, acuerdos, información contractual.
- **Referencias temporales volátiles**: números de ticket, PRs específicos, incidentes con fechas, menciones de releases en curso.
- **Información organizacional**: jerarquía, políticas internas, discusiones no técnicas.

## ⚠️ Zona gris — requiere juicio

- **Inventarios de negocio**: servicios internos descritos por función técnica (OK) vs. por relación con clientes o regulación (no OK).
- **Nombres de proyectos**: cuando son públicos u open-source (OK) vs. cuando son productos internos confidenciales (no OK).
- **Métricas**: performance/throughput sin contexto de negocio (OK) vs. ingresos, usuarios, KPIs (no OK).

Cuando dude, el agente debe **omitir**. Es reversible añadir contenido después; es caro extirparlo de una KG ya distribuida.

---

## Cómo se aplica

- **Al escribir**: el `orchestrator` y cualquier agente que persista al KG deben filtrar el contenido contra esta política antes de llamar a `create_entities` / `add_observations`. Si una observación cae en zona prohibida, se omite silenciosamente; si cae en zona gris, se omite por default.
- **Al exportar**: `export.py` (futuro) confía en que la KG local ya cumple la política — no hace curación propia.
- **Al importar**: `import.py` (futuro) confía en que el archivo origen ya cumple la política — no hace filtrado.

La responsabilidad del filtro vive **una sola vez**, al momento de escribir.

## Qué hacer si detectas una violación

- **En tu KG local**: elimina la entidad/observación con el skill `/memory` o con el viewer (`uv run chromadb-mcp/viewer/app.py`).
- **En un archivo compartido**: rechaza el PR de `shared-knowledge/`, pide al origen que cure y re-exporte.
- **En el código de un agente**: abre un PR ajustando el prompt del agente para que respete la política.

## Convenciones de idioma

Todo el contenido del KG se escribe en **inglés**, independientemente del idioma de la conversación. Esto alinea con la convención existente del sistema y facilita compartir entre equipos.

---

## Estado

**Versión**: 0.1 (borrador inicial, 2026-04-22).

**Estado de implementación**:
- ✅ Filtro cableado en `orchestrator.md` Phase 6 (Knowledge Save).
- ✅ `chromadb-mcp/export.py` y `chromadb-mcp/import.py` disponibles.

Esta política es **normativa para humanos y agentes**. El filtro del orchestrator es la primera línea de defensa al escribir; los humanos curan al revisar PRs en `shared-knowledge/`.
