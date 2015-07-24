import random
from twisted.internet import reactor
import sys
from smap import driver
from smap.util import periodicSequentialCall

from smap.services.zonecontroller import ZoneController

num_floors = 10
num_zones_per_floor = 20

class Fast(driver.SmapDriver):
    def setup(self, opts):
        self.rate = float(opts.get('rate',1))

        self.index = 0
        self.timeseries = []

        for flr in range(1, num_floors):
            for zone in range(1,num_zones_per_floor+1):
                for room in range(1,random.randint(1, 5)+1):
                        print flr, zone, room, flr*1000+zone*10+room
                        path = '/sensors/sensor{0}/temperature'.format(flr*1000+zone*10+room)
                        self.add_timeseries(path,'F',data_type='long')
                        self.timeseries.append({'flr': flr, 'zone': flr*1000+zone*10, 'room': room, 'path': path})

    def addts(self):
        print self.index, len(self.timeseries)
        if self.index == len(self.timeseries):
            return
        x = self.timeseries[self.index]
        flr = x['flr']
        zone = x['zone']
        room = x['room']
        path = x['path']
        self.set_metadata(path, { 'Metadata/Floor': str(flr),
            'Metadata/Room': str(flr*1000 + zone*10 + room),
            'Metadata/HVAC/Zone': str(zone),
            'Metadata/Location': 'Room',
            'Metadata/Sensor/Measure': 'Temperature',
            'Metadata/Sensor/Type': 'Sensor',
        })
        print 'added', path
        self.index += 1
        

    def start(self):
        periodicSequentialCall(self.addts).start(self.rate)
        periodicSequentialCall(self.read_all).start(self.rate)

    def read(self):
        try:
            self.add(self.timeseries[self.index]['path'], 1)
        except Exception as e:
            print e
            self.stop()
            reactor.stop()
            sys.exit(0)
        #for x in self.timeseries[:self.index]:
        #    try:
        #        self.add(x['path'], 1)
        #    except Exception as e:
        #        print e

    def read_all(self):
        for x in self.timeseries[:self.index]:
            try:
                self.add(x['path'], 1)
            except Exception as e:
                print e
