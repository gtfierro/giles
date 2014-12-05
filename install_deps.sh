for package in archiver httphandler wshandler ; do
  echo "Installing go dependencies for $package"
  archiver_packages=`cd $package && go list -f "{{range .Imports}}{{ . }} {{end}} ."`
  for i in $archiver_packages ; do 
      if [ "$i" == "." ] ; then
          break # skip the "self" package
      fi
      go get -v $i
  done
done
