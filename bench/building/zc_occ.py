import random
from smap import driver
from smap.util import periodicSequentialCall
from twisted.internet import task
from twisted.internet import reactor
from smap.archiver.client import RepublishClient
import time

num_floors = 10
num_zones_per_floor = 20

class ZC(driver.SmapDriver):
    def setup(self, opts):
        self.archiver_url = opts.get('archiver','http://localhost:8079')
        self.repubclients = {}

        # one for each zone
        for floor in range(1, num_floors):
            print floor
            for zone in range(1, num_zones_per_floor):
                query = "Metadata/Zone = '{zone}' and Metadata/Sensor/Measure = 'Occupancy'".format(zone=str(floor*1000+zone*10))
                rc = RepublishClient(self.archiver_url, self.cb, restrict=query)
                self.repubclients['zone{0}'.format(floor*1000+zone*10)] = rc
        print 'DONE'

    def cb(self, *args):
        pass

    def start(self):
        for i, c in enumerate(self.repubclients.itervalues()):
            #reactor.callLater(i, c.connect)
            c.connect()
            #time.sleep(1)

    def stop(self):
        for c in self.repubclients.itervalues():
            c.close()

