#!/bin/bash

server_label_id="bitrise.io.tools.cmd-bridge"
curr_user_lib_launch_agents_dir="$HOME/Library/LaunchAgents"
server_plist_path="${curr_user_lib_launch_agents_dir}/${server_label_id}.plist"
server_full_path="/usr/local/bin/cmd-bridge"
server_logs_dir_path="${HOME}/logs"
server_log_file_path="${server_logs_dir_path}/${server_label_id}.log"
