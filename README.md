# Workspacer
Tmux session manager on steroids 💉

# Tmux integration
```
# Prompt to run workspacer with arguments
bind-key P command-prompt -p "Workspacer args:" "new-window 'workspacer -W=current %1 %2 %3'"

# replacing default chose-tree with a filtered one based on the current workspace
bind-key S choose-tree -Zs
bind-key s run-shell "tmux choose-tree -Zs -f \"$(workspacer -W=current get-tmux-workspace-filter)\""
```



# Goals
- Get conveniently open projects within multiple workspaces (personal, work, etc)
- Be able to select workspace by a set alias in shell config 
```
alias space1="workspacer -W=space1" 
```
- Use of TUIs to make project and session selections
- Pull in info from github with github client
    - Be able to pull in different gh accounts based on workspaces (one gh account for work and one personal)



# TODOs
- [ ] implement `new` sub-command
- [ ] implement `clone` sub-command
