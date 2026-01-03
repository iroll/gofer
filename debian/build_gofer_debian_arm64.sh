# Clean, Build for current arch, and move output to debian/
fakeroot debian/rules clean && dpkg-buildpackage -us -uc -b -aarm64 && mv ../gofer_*.deb ./debian/
