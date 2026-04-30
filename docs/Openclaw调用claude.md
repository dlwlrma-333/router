首先安装(需要node22和npm)openclaw和claude

### 2. 一次性创建独立目录和配置

默认放在 `$HOME/openclaw-router-claude`。如果你想放到当前目录，把下面第一行改成 `export OC_ROOT="$PWD/openclaw-router-claude"`。

先把这两个值改成你自己的：

- `ROUTER_BASE_URL`
- `ROUTER_API_KEY`

```bash
export OC_ROOT="$HOME/openclaw-router-claude"
export OC_BASE="$OC_ROOT/.openclaw"
export ROUTER_BASE_URL="https://router.example.com"
export ROUTER_API_KEY="sk-your-router-key"
export OC_GATEWAY_PORT=18840
export OC_GATEWAY_TOKEN="$(openssl rand -hex 24)"

mkdir -p "$OC_BASE/state" "$OC_BASE/workspace" "$OC_BASE/claude"

cat > "$OC_BASE/claude/settings.json" <<EOF
{
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "$ROUTER_API_KEY",
    "ANTHROPIC_BASE_URL": "$ROUTER_BASE_URL",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1"
  },
  "permissions": {
    "allow": [],
    "deny": []
  }
}
EOF

cat > "$OC_BASE/claude/openclaw-settings.json" <<'EOF'
{
  "permissions": {
    "allow": [
      "Skill",
      "Task",
      "TodoWrite",
      "Read",
      "Glob",
      "Grep",
      "WebFetch",
      "WebSearch",
      "Bash(curl:*)",
      "Bash(wget:*)",
      "Bash(rg:*)",
      "Bash(find:*)",
      "Bash(ls:*)",
      "Bash(cat:*)",
      "Bash(head:*)",
      "Bash(tail:*)",
      "Bash(sed:*)",
      "Bash(awk:*)",
      "Bash(jq:*)",
      "Bash(date:*)",
      "Bash(env:*)",
      "Bash(printenv:*)",
      "Bash(which:*)",
      "Bash(python3:*)"
    ],
    "deny": []
  }
}
EOF

cat > "$OC_BASE/openclaw.json" <<EOF
{
  "meta": {
    "lastTouchedVersion": "2026.4.15"
  },
  "agents": {
    "defaults": {
      "workspace": "$OC_BASE/workspace",
      "model": {
        "primary": "claude-cli/claude-sonnet-4-6"
      },
      "timeoutSeconds": 1800,
      "compaction": {
        "mode": "safeguard"
      },
      "cliBackends": {
        "claude-cli": {
          "command": "claude",
          "args": [
            "-p",
            "--output-format",
            "stream-json",
            "--include-partial-messages",
            "--verbose",
            "--setting-sources",
            "user",
            "--settings",
            "$OC_BASE/claude/openclaw-settings.json",
            "--permission-mode",
            "dontAsk"
          ],
          "env": {
            "CLAUDE_CONFIG_DIR": "$OC_BASE/claude"
          }
        }
      }
    }
  },
  "gateway": {
    "port": $OC_GATEWAY_PORT,
    "mode": "local",
    "auth": {
      "mode": "token",
      "token": "$OC_GATEWAY_TOKEN"
    }
  }
}
EOF

cat > "$OC_ROOT/openclaw-router" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_DIR="$ROOT_DIR/.openclaw"

export OPENCLAW_STATE_DIR="$BASE_DIR/state"
export OPENCLAW_CONFIG_PATH="$BASE_DIR/openclaw.json"

exec openclaw "$@"
EOF

chmod +x "$OC_ROOT/openclaw-router"
```
### 3.测试

```bash
cd "$OC_ROOT"

timeout 180 ./openclaw-router agent --local \
  --session-id smoke \
  --message 'Reply with OPENCLAW_READY only.'

echo "exit=$?"
```
看到回`OPENCLAW_READY`就是成功了
