cd giles
rm giles-*
PLATFORMS="darwin/386 darwin/amd64 linux/386 linux/amd64"
for platform in $PLATFORMS ; do
GOOS=${platform%/*}
GOARCH=${platform#*/}
echo "building $GOOS-$GOARCH"
go build
mv giles giles-$GOOS-$GOARCH
done
cd ..
