# -*- coding: utf-8 -*-
"""Caliopen NATS listener for message processing."""
from __future__ import absolute_import, print_function, unicode_literals

import argparse
import sys
import logging

import tornado.ioloop
import tornado.gen
from nats.io import Client as Nats

from caliopen_storage.config import Configuration
from caliopen_storage.helpers.connection import connect_storage

log = logging.getLogger(__name__)
logging.basicConfig(level=logging.INFO)


@tornado.gen.coroutine
def inbound_smtp_handler(config):
    """Inbound message NATS handler."""
    client = Nats()
    server = 'nats://{}:{}'.format(config['host'], config['port'])
    servers = [server]

    opts = {"servers": servers}
    yield client.connect(**opts)

    # create and register subscriber(s)
    inbound_email_sub = subscribers.InboundEmail(client)
    future = client.subscribe("inboundSMTP", "SMTPqueue",
                              inbound_email_sub.handler)
    log.info("nats subscription started for inboundSMTP")
    future.result()

@tornado.gen.coroutine
def inbound_twitter_handler(config):
    """Inbound twitterDM NATS handler"""
    client = Nats()
    server = 'nats://{}:{}'.format(config['host'], config['port'])
    servers = [server]

    opts = {"servers": servers}
    yield client.connect(**opts)

    # create and register subscriber(s)
    inbound_twitter_sub = subscribers.InboundTwitter(client)
    future = client.subscribe("inboundTwitter", "Twitterqueue",
                              inbound_twitter_sub.handler)
    log.info("nats subscription started for inboundTwitter")
    future.result()

@tornado.gen.coroutine
def inbound_mastodon_handler(config):
    """Inbound mastodonDM NATS handler"""
    client = Nats()
    server = 'nats://{}:{}'.format(config['host'], config['port'])
    servers = [server]

    opts = {"servers": servers}
    yield client.connect(**opts)

    # create and register subscriber(s)
    inbound_mastodon_sub = subscribers.InboundMastodon(client)
    future = client.subscribe("inboundMastodon", "Mastodonqueue",
                              inbound_mastodon_sub.handler)
    log.info("nats subscription started for inboundMastodon")
    future.result()

@tornado.gen.coroutine
def contact_handler(config):
    """NATS handler for contact update events."""
    client = Nats()
    server = 'nats://{}:{}'.format(config['host'], config['port'])
    servers = [server]

    opts = {"servers": servers}
    yield client.connect(**opts)

    # create and register subscriber(s)
    contact_subscriber = subscribers.ContactAction(client)
    future = client.subscribe("contactAction", "contactQueue",
                     contact_subscriber.handler)
    log.info("nats subscription started for contactAction")
    future.result()


@tornado.gen.coroutine
def key_handler(config):
    """NATS handler for discover_key events."""
    client = Nats()
    server = 'nats://{}:{}'.format(config['host'], config['port'])
    servers = [server]

    opts = {"servers": servers}
    yield client.connect(**opts)

    # create and register subscriber(s)
    key_subscriber = subscribers.KeyAction(client)
    future = client.subscribe("keyAction",
                              "keyQueue",
                              key_subscriber.handler)
    log.info("nats subscription started for keyAction")
    future.result()


if __name__ == '__main__':
    # load Caliopen config
    args = sys.argv
    parser = argparse.ArgumentParser()
    parser.add_argument('-f', dest='conffile')
    kwargs = parser.parse_args(args[1:])
    kwargs = vars(kwargs)
    Configuration.load(kwargs['conffile'], 'global')
    # need to load config before importing subscribers
    import subscribers

    connect_storage()
    inbound_smtp_handler(Configuration('global').get('message_queue'))
    inbound_twitter_handler(Configuration('global').get('message_queue'))
    inbound_mastodon_handler(Configuration('global').get('message_queue'))
    contact_handler(Configuration('global').get('message_queue'))
    key_handler(Configuration('global').get('message_queue'))
    loop_instance = tornado.ioloop.IOLoop.instance()
    loop_instance.start()
