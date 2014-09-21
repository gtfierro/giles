import subprocess
import sys
import glob
import os
FNULL = open(os.devnull, 'w')
inis = glob.glob('*.ini')
inis = [x.split('.')[0] for x in inis]

print inis

processes = []
for ini in inis:
    command = "twistd --pidfile {0}.pid -n smap {0}.ini".format(ini)
    p = subprocess.Popen(command, shell=True, stdout=FNULL)
    processes.append(p)
# run smap query to see if the streams are registered
# This is partly to get around the race condition arising when
# a driver subscribes to the output of another
# However, top handle the mutual dependence case, it really needs to
# restart the services
# Note that this encodes the uri of the archiver
try:
    raw_input()
except:
    for p in processes:
        print "killing",p
        p.terminate()
