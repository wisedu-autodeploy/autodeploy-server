rm -rf bin/darwin bin/linux bin/windows &&
mkdir -p bin/darwin bin/linux bin/windows &&

go-bindata-assetfs -o static.go web/... &&

gox -osarch="linux/amd64 windows/amd64 darwin/amd64" -output="autodeploy_{{.OS}}_{{.Arch}}" &&

mv autodeploy_darwin* bin/darwin/autodeploy &&
# cp -r web bin/darwin/web &&

mv autodeploy_linux* bin/linux/autodeploy &&
# cp -r web bin/linux/web &&

mv autodeploy_windows* bin/windows/autodeploy.exe &&
# cp -r web bin/windows/web &&

chmod +x bin/darwin/autodeploy