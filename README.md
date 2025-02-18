uploader cli and server for https://dj.cansu.dev

wip

### install

only dependency is uv, `uvx` is not needed, `uv` must be in system path. skip if already installed.

<details>

<summary> uv </summary>

macos // linux
```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
# inspect the script if you want to
curl -LsSf https://astral.sh/uv/install.sh | less
```
windows
```bash
powershell -ExecutionPolicy ByPass -c "irm https://astral.sh/uv/install.ps1 | iex"
# inspect the script if you want to
powershell -c "irm https://astral.sh/uv/install.ps1 | more"
```
if you dont want installation scripts just https://docs.astral.sh/uv/getting-started/installation/#installation-methods pick one, make sure `uv` binary is in path.


</details>

then:

```bash
# if you want to build for your own system
just build-current
# will yield
# dist
# └── strafe
#
# if you want to build for all systems
just build
# will yield
# dist
# ├── strafe-darwin-amd64
# ├── strafe-darwin-arm64
# ├── strafe-linux-amd64
# ├── strafe-linux-arm64
# ├── strafe-windows-amd64.exe
# └── strafe-windows-arm64.exe
```