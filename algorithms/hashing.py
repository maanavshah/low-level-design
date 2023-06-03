class Hash:
    def __init__(self, size=10):
        self.size = size
        self.data = [[] for _ in range(self.size)]

    def _hash(self, key):
        return hash(key) % self.size

    def put(self, key, value):
        bucket = self._hash(key)
        for node in self.data[bucket]:
            if node[0] == key:
                node[1] = value
                return
        self.data[bucket].append([key, value])

    def get(self, key):
        bucket = self._hash(key)
        for node in self.data[bucket]:
            if node[0] == key:
                return node[1]
        return None

    def remove(self, key):
        bucket = self._hash(key)
        data = self.data[bucket]
        for idx, node in enumerate(data):
            if node[0] == key:
                data.pop(idx)
                return
