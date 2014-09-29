from smap import driver
from smap.util import periodicSequentialCall

class Fast(driver.SmapDriver):
    def setup(self, opts):
        self.rate = float(opts.get('rate',.1))
        self.number = int(opts.get('number',5))

        self.values = [0]*self.number

        for i in range(self.number):
            self.add_timeseries('/sensor{0}'.format(i),'V',data_type='long')

    def start(self):
        periodicSequentialCall(self.read).start(self.rate)

    def read(self):
        for i in range(self.number):
            self.values[i] += 1
            self.add('/sensor{0}'.format(i), self.values[i])

