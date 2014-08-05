from smap import driver
from smap.util import periodicSequentialCall

class Fast(driver.SmapDriver):
    def setup(self, opts):
        self.rate = float(opts.get('rate',.1))

        self.values = [0]*5

        for i in range(5):
            self.add_timeseries('/sensor{0}'.format(i),'V',data_type='long')

    def start(self):
        periodicSequentialCall(self.read).start(self.rate)

    def read(self):
        for i in range(5):
            self.values[i] += 1
            self.add('/sensor{0}'.format(i), self.values[i])

