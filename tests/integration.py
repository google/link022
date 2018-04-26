"""
Integration tests for link022 gnmi functions.

Refer to the README.md file for instructions on running the tests.
"""

import argparse
import json
import netaddr
import logging
import platform
import sys
import time
import tempfile
import unittest

import mininet.net
import mininet.node
import mininet.cli


FLAGS = None
logging.basicConfig(stream=sys.stdout, level=logging.INFO)
logger = logging.getLogger()


def set_flags():
    """Set the global FLAGS
    """
    global FLAGS
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--target_cmd",
        help="Command line to start the target. Needed if ext_target is False")
    parser.add_argument(
        "--ext_target", action='store_true', default=False,
        help='Use an external target.')
    parser.add_argument(
        "--emulator", action='store_true', default=False,
        help='Start the emulator and run no tests.')
    parser.add_argument("--gnmi_set", help="Path to the gnmi_set command")
    parser.add_argument("--ca", help="CA Certificate")
    parser.add_argument("--cert", help="Client Certificate")
    parser.add_argument("--key", help="Client key")
    parser.add_argument(
        "--target_name", help="Target name for cert verification")
    parser.add_argument("--target_addr", help="Target IP:port")
    parser.add_argument("--json_conf", help="File name of JSON config")
    FLAGS = parser.parse_args()


TARGET_NAME = 'target'
CONTROLLER_NAME = 'ctrlr'
DUMMY_NAME = 'dummy'
DEFAULT_NS = 'lk022_def'


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
        cls._target_popen = None
        if FLAGS.ext_target:
            cls._ctrlr = mininet.node.Node(DEFAULT_NS, inNamespace=False)
        else:
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
        logger.info('Running target command: %s', FLAGS.target_cmd)
        cls._target_popen = cls._target.popen(FLAGS.target_cmd)

        if FLAGS.emulator:
            mininet.cli.CLI(cls._net)
        else:
            # Wait for the agent to start up
            time.sleep(20)

    @classmethod
    def tearDownClass(cls):
        """Test clean up.
        """
        if cls._target_popen:
            cls._target_popen.kill()
            cls._target_popen = None

    def runTest(self):
        """Run the test
        """
        if FLAGS.emulator:
            self.skipTest('Skip the test in emulator mode.')
        ap_conf_file = FLAGS.json_conf
        if not FLAGS.ext_target:
            # Rewrite the hostname for emulator based tests.
            ap_conf = json.load(open(ap_conf_file, 'r'))
            for ap in ap_conf['openconfig-access-points:access-points']['access-point']:
                ap['hostname'] = platform.node()
            json_h = tempfile.NamedTemporaryFile()
            json.dump(ap_conf, json_h)
            json_h.flush()
            ap_conf_file = json_h.name
        gnmi_set_cmd_list = (
            FLAGS.gnmi_set,
            '-ca=' + FLAGS.ca,
            '-cert=' + FLAGS.cert,
            '-key=' + FLAGS.key,
            '-target_name=' + FLAGS.target_name,
            '-target_addr=' + FLAGS.target_addr,
            '-replace=' + '/:@' + ap_conf_file)
        gnmi_set_cmd = ' '.join(gnmi_set_cmd_list)

        logger.info('Running gnmi_set command: %s', gnmi_set_cmd)
        _, _, code = self._ctrlr.pexec(gnmi_set_cmd)
        self.assertEqual(code, 0)
        if not FLAGS.ext_target:
            json_h.close()


if __name__ == '__main__':
    set_flags()
    tests = unittest.TestSuite()
    tests.addTest(ConfigTest())
    unittest.TextTestRunner().run(tests)
