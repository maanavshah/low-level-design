import hashlib

HEXADECIMAL_BASE = 16


class ConsistentHashing:
    def __init__(self, nodes=None, virtual_nodes=3):
        self.virtual_nodes = virtual_nodes
        self.ring = {}
        self.sorted_keys = []

        if nodes:
            for node in nodes:
                self.add_node(node)

    def add_node(self, node):
        for i in range(self.virtual_nodes):
            key = self._get_hash("{}-{}".format(node, i))
            self.ring[key] = node
            self.sorted_keys.append(key)

        self.sorted_keys.sort()

    def remove_node(self, node):
        for i in range(self.virtual_nodes):
            key = self._get_hash("{}-{}".format(node, i))
            del self.ring[key]
            self.sorted_keys.remove(key)

    def get_node(self, key):
        if not self.ring:
            return None

        hash_key = self._get_hash(key)
        for ring_key in self.sorted_keys:
            if hash_key <= ring_key:
                return self.ring[ring_key]

        # Wrap around to the first node in the ring
        return self.ring[self.sorted_keys[0]]

    def _get_hash(self, value):
        return int(hashlib.md5(value.encode()).hexdigest(), HEXADECIMAL_BASE)
