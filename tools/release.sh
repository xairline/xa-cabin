rm -rf dist
npx concurrently "nx run xws:build" "cd apps/core; make all -j 3"
rm -f dist/apps/XWebStack/mac_arm.xpl dist/apps/XWebStack/mac_amd.xpl
cp -r dist/apps/xws dist/apps/XWebStack
echo "PORT=19090" > dist/apps/XWebStack/config
root=$(pwd)
cd dist/apps/ && zip -r xws.zip XWebStack && cd $root
