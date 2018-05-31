from __future__ import absolute_import
from __future__ import division
from __future__ import print_function
from __future__ import unicode_literals

import argparse
import logging
import warnings
import spacy
from spacy.cli import download
from policy import RestaurantPolicy
from rasa_core import utils, server
from rasa_core.actions import Action
from rasa_core.agent import Agent
from rasa_core.channels.console import ConsoleInputChannel
from rasa_core.events import SlotSet
from rasa_core.featurizers import (
    MaxHistoryTrackerFeaturizer,
    BinarySingleStateFeaturizer)
from rasa_core.interpreter import RasaNLUInterpreter
from rasa_core.policies.memoization import MemoizationPolicy
from gevent.pywsgi import WSGIServer

logger = logging.getLogger(__name__)


class RestaurantAPI(object):
    def search(self, info):
        return "papi's pizza place"


class ActionSearchRestaurants(Action):
    def name(self):
        return 'action_search_restaurants'

    def run(self, dispatcher, tracker, domain):
        dispatcher.utter_message("looking for restaurants")
        restaurant_api = RestaurantAPI()
        restaurants = restaurant_api.search(tracker.get_slot("cuisine"))
        return [SlotSet("matches", restaurants)]

def run(serve_forever=True):
    download("en")
    download("de")
    interpreter = RasaNLUInterpreter("models/current/nlu")
    bot = server.create_app("models/dialogue", interpreter=interpreter)

    logger.info("Started http server on port %s" % '8080')

    http_server = WSGIServer(('0.0.0.0', 8080), bot)
    logger.info("Up and running")
    try:
        http_server.serve_forever()
    except Exception as exc:
        logger.exception(exc)

    return bot


if __name__ == '__main__':
    utils.configure_colored_logging(loglevel="INFO")
    run()