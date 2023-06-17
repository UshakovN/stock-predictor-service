import http
import time
import os
import ml

import charts
import schedule
import pandas as pd
import requests as req

from datetime import datetime
from dotenv import load_dotenv
from loguru import logger as log
from threading import Thread


_API_TOKEN_HEADER = 'X-Service-Token'
_AUTH_TOKEN_HEADER = 'X-Auth-Token'

_API_TOKEN_STOCKS = 'API_TOKEN_STOCKS'
_API_ROUTE_STOCKS = 'API_ROUTE_STOCKS'

_API_ROUTE_AUTH = 'API_ROUTE_AUTH'
_SERVICE_HASH_SALT = 'SERVICE_HASH_SALT'


class ForbiddenException(Exception):
    pass


class UnauthorizedException(Exception):
    pass


class Service:
    __api_token_stocks: str
    __api_route_stocks: str
    __api_route_auth: str
    __charts_builder: charts.ChartsBuilder
    __price_movement_predictor: ml.PriceMovementPredictor

    def load_config(self, config_path: str):
        load_dotenv(config_path)
        api_token_stocks = os.getenv(_API_TOKEN_STOCKS)
        api_route_stocks = os.getenv(_API_ROUTE_STOCKS)
        api_route_auth = os.getenv(_API_ROUTE_AUTH)
        service_hash_salt = os.getenv(_SERVICE_HASH_SALT)

        for field in [api_token_stocks, api_route_stocks, api_route_auth, service_hash_salt]:
            if field in [None, ""]:
                raise Exception(f"some required fields not found in config {config_path}")

        self.__api_token_stocks = api_token_stocks
        self.__api_route_stocks = api_route_stocks
        self.__api_route_auth = api_route_auth

        self.__charts_builder = charts.ChartsBuilder(service_hash_salt)
        self.__price_movement_predictor = ml.PriceMovementPredictor(config_path)

    def __load_stocks_dataframe(self) -> pd.DataFrame:
        reties = 5
        wait = 5
        for retry in range(reties):
            resp = req.get(self.__api_route_stocks, headers={
                _API_TOKEN_HEADER: self.__api_token_stocks,
            })
            if resp.status_code in [http.HTTPStatus.FORBIDDEN, http.HTTPStatus.UNAUTHORIZED]:
                raise Exception(f"config api token malformed or expired. specified header: {_API_TOKEN_HEADER}")

            if resp.status_code == http.HTTPStatus.NOT_FOUND:
                raise Exception(f"malformed request api route: {self.__api_route_stocks}")

            if resp.status_code >= http.HTTPStatus.INTERNAL_SERVER_ERROR:
                log.warning(f"cannot get stocks from {self.__api_route_stocks}. got status code: {resp.status_code}")
                reties -= 1
                time.sleep(wait)
                continue

            content = resp.json()

            if not bool(content['success']):
                log.warning(f"stocks response status code is not success. got: {content.success}")

            if (content or content['stocks']) is None or len(content['stocks']) == 0:
                raise Exception(f"stocks json response from {self.__api_route_stocks} is a none")

            df = pd.DataFrame(content['stocks'])
            return df

    def __update_service_components(self):
        df = self.__load_stocks_dataframe()
        log.info("service loaded stocks dataframe")

        self.__charts_builder.set_dataframe(df.copy())
        log.info("charts builder set loaded dataframe")

        self.__price_movement_predictor.update_stored_predicts(df.copy())
        log.info("price movement predictor update stored predicts")

    def update_service_components(self):
        self.__update_service_components()
        log.info("first update service components success")

        def update():
            try:
                self.__update_service_components()
                log.info("scheduled update service components success")

            except Exception as ex:
                log.error(f"update service components failed: {ex}")

        schedule.every().day.at("00:00").hours.do(update)

        def scheduler():
            wait_one_day = 1
            wait_seconds = wait_one_day * 24 * 60
            time.sleep(wait_seconds)

            while True:
                schedule.run_pending()

        log.info("run scheduled update service components")
        th = Thread(target=scheduler, args=())
        th.start()

    def create_chart(
            self,
            ticker_id: str,
            from_date: datetime,
            to_date: datetime,
            indicators: bool,
            force_refresh: bool,
    ) -> str:
        return self.__charts_builder.create_chart(ticker_id, from_date, to_date, indicators, force_refresh)

    def check_client(self, auth_token: str):
        resp: req.Response
        try:
            resp = req.get(self.__api_route_auth, headers={
                _AUTH_TOKEN_HEADER: auth_token,
            })
        except req.RequestException as ex:
            raise Exception(f"request failed: {ex}")

        if resp.status_code == http.HTTPStatus.OK:
            return

        content = resp.json()
        if content is None:
            raise Exception("response content not found")

        message = content['message']
        if message in [None, ""]:
            message = "forbidden"

        if resp.status_code == http.HTTPStatus.UNAUTHORIZED:
            raise UnauthorizedException(message)

        if resp.status_code == http.HTTPStatus.FORBIDDEN:
            raise ForbiddenException(message)

        raise Exception(f"unexpected response status code {resp.status_code}")





