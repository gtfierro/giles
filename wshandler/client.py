

from ws4py.client.threadedclient import WebSocketClient

class DummyClient(WebSocketClient):
    def opened(self):

        self.send('Metadata/Type = "Command"')

    def closed(self, code, reason=None):
        print "Closed down", code, reason

    def received_message(self, m):
        print m

if __name__ == '__main__':
    try:
        ws = DummyClient('ws://localhost:8078/republish')
        ws.connect()
        ws.run_forever()
    except KeyboardInterrupt:
        ws.close()
