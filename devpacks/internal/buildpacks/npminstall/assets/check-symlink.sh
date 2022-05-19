if [ ! -e "${WORKSPACE_FOLDER}/node_modules" ]; then
    ln -s "${NPM_INSTALL_LAYER}/node_modules" "${WORKSPACE_FOLDER}/node_modules"
fi