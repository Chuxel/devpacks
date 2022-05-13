#!/bin/bash

# The CNB_APP_DIR is filtered out by the launcher, so this is an easy way to detect if the current process
# was started via the launcher. In the event of an "docker exec", the containe entrypoint is not fired, so
# the shell process will not be launched properly. So, what this will do is detect if we're in a login or 
# interactive sh, bash, or zsh process and replace the shell with one started using /cnb/lifecycle/launcher
# with the exact same arguments. In devcontainer.json, we'll then force userEnvProbe to be loginInteractiveShell.
#
# More info on launcher environment: https://github.com/buildpacks/spec/blob/main/platform.md#launch-environment
snippet="$(cat << 'EOF'
# Ensure all interactive or login shells are initalized via /cnb/lifecycle/launcher (which is also the default entrypoint)
if [ ! -z "${CNB_APP_DIR}" ] && [ -z "${DCNB_ENV_LOADED}" ] && [ -e "/cnb/lifecycle/launcher" ]; then export DCNB_ENV_LOADED=true; mapfile -d $'\0' _cmd_line < /proc/$$/cmdline; exec /cnb/lifecycle/launcher "${_cmd_line[@]//\"/\\\"}"; fi
EOF
)"

add_to_top_of_file() {
    local filename="$1"
    local check_exists="${2:-$1}"
    if [ ! -e "${check_exists}" ]; then
        echo "${check_exists} does not exist. Skipping."
        return
    fi
    local existing_file="$(cat "${filename}" 2>/dev/null)"
    if ! grep -Fxq "${snippet//\0/\\0}" "${filename}" 2>/dev/null; then
        echo "${snippet}
${existing_file}" > "${filename}"
        echo "Adding /cnb/lifecycle/launcher to ${filename}."
    else 
        echo "/cnb/lifecycle/launcher already exists in ${filename}. Skipping."
    fi
}

add_to_top_of_file /etc/bash.bashrc
add_to_top_of_file /etc/profile
add_to_top_of_file /etc/zsh/zshenv /etc/zsh
