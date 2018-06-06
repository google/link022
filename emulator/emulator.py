"""Emulator for link022. Refer to the README.md file for instructions.
"""

import argparse
import netaddr
import logging

import mininet.net
import mininet.node
import mininet.cli


FLAGS = None
logging.basicConfig(filename='/tmp/link022_emulator.log', level=logging.INFO)
logger = logging.getLogger()


def set_flags():
  """Set the global FLAGS."""
  global FLAGS
  parser = argparse.ArgumentParser()
  parser.add_argument(
      '--target_cmd',
      help='Command line to start the target.')
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


class Emulator(object):
  """Link022 emulator."""

  def __init__(self):
    self._net = None
    self._ctrlr = None
    self._target = None
    self._target_popen = None

  def start(self):
    self._start_topo()
    logger.info('Emulator started.')
    mininet.cli.CLI(self._net)

  def _start_topo(self):
    """Create an empty network and add nodes to it.
    """

    subnet = netaddr.IPNetwork('10.0.0.0/24')
    hosts_iter = subnet.iter_hosts()
    self._net = mininet.net.Mininet(controller=None)

    self._net.addHost(TARGET_NAME)
    self._net.addHost(CONTROLLER_NAME)
    # We add a dummy host to create the eth and wlan interfaces for the
    # Target
    self._net.addHost(DUMMY_NAME)
    params1 = {'ip': get_ip_spec(hosts_iter.next(), subnet)}
    params2 = {'ip': get_ip_spec(hosts_iter.next(), subnet)}
    self._net.addLink(TARGET_NAME, CONTROLLER_NAME,
                      params1=params1, params2=params2)
    self._net.addLink(TARGET_NAME, DUMMY_NAME)
    self._net.addLink(TARGET_NAME, DUMMY_NAME)

    self._target = self._net[TARGET_NAME]
    self._ctrlr = self._net[CONTROLLER_NAME]

    self._net.start()
    logger.info('Running Link022 target command: %s', FLAGS.target_cmd)
    self._target_popen = self._target.popen(FLAGS.target_cmd)

  def cleanup(self):
    """Clean up emulator."""
    if self._target_popen:
      self._target_popen.kill()
      self._target_popen = None
    if self._net:
      self._net.stop()
      logger.info('Emulator cleaned up.')

if __name__ == '__main__':
  set_flags()
  emulator = Emulator()
  try:
    emulator.start()
  finally:
    emulator.cleanup()

