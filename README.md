# git-mirror

This little go service receives GitHub webhooks, pulls a mirror of
a git repository, and pushes it to another remote, for example, a
GitLab install.

It's small enough to audit, easy enough to write a SysV service file
for, or a SystemD service file, and you don't need to buy GitLab EE!
