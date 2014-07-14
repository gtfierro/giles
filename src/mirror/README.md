Sometimes, we may not want to setup a full sMAP archiver, and instead just have
a central location that all sMAP sources can report to such that processes can
receive streaming data from any sMAP source that is reporting to the
republisher. This functionality will PROBABLY be integrated into the archiver,
if it doesn't exist already.

This is just a dumb exercise for something to do on the plane.

localhost:8079/republish/UUIDHERE

or i guess you could publish a dictionary of key-value pairs and it would 'subscribe'
you to all the points that have those tags.
