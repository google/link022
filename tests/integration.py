"""
Integration tests for link022 gnmi functions.

Refer to the README.md file for instructions on running the tests.
"""

import argparse
import netaddr
import time
import unittest

import mininet.net
import mininet.node
import mininet.cli


FLAGS = None


def set_flags():
    """Set the global FLAGS
    """
    global FLAGS
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--target_cmd", help="Command line to start the target")
    parser.add_argument("--gnmi_set", help="Command line for GNMI Set")
    FLAGS = parser.parse_args()


TARGET_NAME = 'target'
CONTROLLER_NAME = 'ctrlr'
DUMMY_NAME = 'dummy'


def get_ip_spec(addr, subnet=None):
    """Get the IP address with the subnet prefix length.

    Args:
      addr: network.addr object
      subnet: network.net object

    Returns:
      A string of the IP address with prefix length.

    Raises:
      Exception: if ip not in subnet
    """
    if subnet is None:
        if addr.version == 6:
            ip_spec = str(addr) + '/128'
        else:
            ip_spec = str(addr) + '/32'
    elif addr in subnet:
        ip_spec = '%s/%s' % (addr, subnet.prefixlen)
    else:
        raise Exception('ip %s is not in subnet %s' % (addr, subnet))
    return ip_spec


class ConfigTest(unittest.TestCase):
    """Define the test driver.
    """
    @classmethod
    def setUpClass(cls):
        cls._start_topo()

    @classmethod
    def _start_topo(cls):
        """Create an empty network and add nodes to it.
        """

        subnet = netaddr.IPNetwork('10.0.0.0/24')
        hosts_iter = subnet.iter_hosts()
        cls._net = mininet.net.Mininet(controller=None)

        cls._net.addHost(TARGET_NAME)
        cls._net.addHost(CONTROLLER_NAME)
        # We add a dummy host to create the eth and wlan interfaces for the
        # Target
        cls._net.addHost(DUMMY_NAME)
        params1 = {'ip': get_ip_spec(hosts_iter.next(), subnet)}
        params2 = {'ip': get_ip_spec(hosts_iter.next(), subnet)}
        cls._net.addLink(TARGET_NAME, CONTROLLER_NAME,
                         params1=params1, params2=params2)
        cls._net.addLink(TARGET_NAME, DUMMY_NAME)
        cls._net.addLink(TARGET_NAME, DUMMY_NAME)

        cls._target = cls._net[TARGET_NAME]
        cls._ctrlr = cls._net[CONTROLLER_NAME]

        cls._net.start()
        cls._target_popen = cls._target.popen(FLAGS.target_cmd)
        # Wait for the agent to start up
        time.sleep(20)

    @classmethod
    def tearDownClass(cls):
        """Test clean up.
        """
        cls._target_popen.kill()

    def runTest(self):
        """Run the test
        """
        _, _, code = self._ctrlr.pexec(FLAGS.gnmi_set)
        self.assertEqual(code, 0)



if __name__ == '__main__':
    set_flags()
    tests = unittest.TestSuite()
    tests.addTest(ConfigTest())
    unittest.TextTestRunner().run(tests)
