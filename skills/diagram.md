Create an Excalidraw diagram that visually argues a concept, workflow, or architecture. Routes through the orchestrator which delegates to the excalidraw-diagram skill.

## Input

$ARGUMENTS — describe what to diagram. Examples:
- "the dev-team pipeline flow"
- "authentication flow for the login system"
- "how the orchestrator delegates to agents"
- A topic without description → the orchestrator infers what to visualize

## What happens

1. Pass to the `orchestrator` agent:

```
Direct Mode Task: diagram
Description: {$ARGUMENTS}
Output: {path where .excalidraw file should be saved, or "ask the user"}
```

## Rules

- Always invoke the `orchestrator` agent — do NOT invoke the excalidraw-diagram skill directly
- The orchestrator will load the excalidraw-diagram skill context and generate the diagram
- The skill handles render validation (Playwright render loop) internally
- Output is a `.excalidraw` file (and optionally a PNG preview)
