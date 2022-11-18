rm .\.git\index.lock -Force
git status
git add .
git commit -m "More changes"
git push origin ppereira-prometheus -f