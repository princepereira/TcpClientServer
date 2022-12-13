rm .\.git\index.lock -Force -ErrorAction Ignore
git status
git add .
git commit -m "More changes"
git push -f