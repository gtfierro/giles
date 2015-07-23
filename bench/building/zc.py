import random
from smap import driver
from smap.util import periodicSequentialCall
from smap.archiver.client import RepublishClient

num_floors = 10
num_zones_per_floor = 50

class ZC(driver.SmapDriver):
    def setup(self, opts):
        self.archiver_url = opts.get('archiver','http://localhost:8079')
        self.repubclients = {}

        # one for each floor
        for floor in range(1, num_floors):
            query = "Metadata/Building = 'UC Berkeley' and Metadata/Floor = '{floor}' and Metadata/Sensor/Measure = 'Temperature'"
            rc = RepublishClient(self.archiver_url, self.cb, restrict=query)
            self.repubclients['floor{0}'.format(floor)] = rc

    def cb(self, *args):
        pass

    def start(self):
        for c in self.repubclients.itervalues():
            c.connect()

    def stop(self):
        for c in self.repubclients.itervalues():
            c.close()
        
