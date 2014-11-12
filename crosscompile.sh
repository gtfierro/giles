rm giles-*
PLATFORMS="darwin/386 darwin/amd64 linux/386 linux/amd64"
for platform in $PLATFORMS ; do
GOOS=${platform%/*}
GOARCH=${platform#*/}
echo "building $GOOS-$GOARCH"
go build -o archiver
mv archiver archiver-$GOOS-$GOARCH
done
