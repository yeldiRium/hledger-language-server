#!/usr/bin/env bash
set -eu

session="ide-hledger-language-server"
attach() {
  if [ -n "${TMUX:-}" ]; then
    tmux switch-client -t "=${session}"
  else
    tmux attach-session -t "=${session}"
  fi
}

if ! tmux has-session -t "${session}" 2>/dev/null; then
  tmux new-session -d -s "${session}" -x "$(tput cols)" -y "$(tput lines)"

  tmux split-window -t "${session}:0.0" -v
  tmux resize-pane -t "${session}:0.0" -y "30%"

  sleep 1

  tmux send-keys -t "${session}:0.0" "devbox shell" "C-m"
  tmux send-keys -t "${session}:0.1" "devbox shell" "C-m"

  tmux send-keys -t "${session}:0.0" "nvim" "C-m"

  tmux select-window -t "${session}:0"
fi

attach
