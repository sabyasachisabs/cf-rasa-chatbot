from rasa_core.actions.action import Action
from rasa_core.actions.forms import FormAction
from rasa_core.actions.forms import EntityFormField
from rasa_core.events import SlotSet, AllSlotsReset

import requests
import ldap3
import json
import os
import logging
import sys

logging.basicConfig(level=logging.DEBUG)


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