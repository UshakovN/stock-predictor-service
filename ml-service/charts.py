import pandas as pd
import plotly
import cufflinks as cf

from os import path
from datetime import datetime
from threading import Lock

import tools

setattr(plotly.offline, "__PLOTLY_OFFLINE_INITIALIZED", True)

_SOURCE_DATE_FORMAT = '%Y-%m-%dT%H:%M:%SZ'
_NORMALIZED_DATE_FORMAT = '%Y-%m-%d'

_CHART_BASE_TITLE = '{}: stock price from {} to {}'
_CHART_BASE_PATH = '{}_{}_{}_indicators_{}'


class ChartsBuilder:
    __fixed_df: pd.DataFrame
    __hash_salt: str
    __lock: Lock

    def __init__(self, hash_salt: str):
        self.__hash_salt = hash_salt
        self.__lock = Lock()

    def set_dataframe(self, df: pd.DataFrame):
        self.__lock.acquire()
        self.__fixed_df = df
        self.__lock.release()

    def __get_dataframe(self) -> pd.DataFrame:
        self.__lock.acquire()
        df = self.__fixed_df
        self.__lock.release()
        return df

    def __form_normalized_df(self, ticker_id: str) -> pd.DataFrame:
        df = self.__get_dataframe()
        normalized_df = df.loc[df['ticker_id'] == ticker_id]

        normalized_df.rename(
            columns={
                'open_price': 'open',
                'close_price': 'close',
                'highest_price': 'high',
                'lowest_price': 'low',
                'trading_volume': 'volume'
            },
            inplace=True,
        )
        normalized_df['stocked_time'] = pd.to_datetime(
            normalized_df['stocked_time'],
            format=_SOURCE_DATE_FORMAT,
        )
        normalized_df['date'] = normalized_df['stocked_time'].dt.strftime(_NORMALIZED_DATE_FORMAT)
        normalized_df = normalized_df.set_index('date')

        return normalized_df

    def __form_chart_filename(
            self,
            ticker_id: str,
            from_date: datetime,
            to_date: datetime,
            indicators: bool,
            prefix: bool = False,
            extension: bool = False,
    ) -> str:
        indicators_suffix = "on" if indicators else "off"

        chart_name = _CHART_BASE_PATH.format(
            ticker_id,
            from_date.strftime(_NORMALIZED_DATE_FORMAT),
            to_date.strftime(_NORMALIZED_DATE_FORMAT),
            indicators_suffix,
        )
        chart_name = tools.create_hash(chart_name, self.__hash_salt)

        if prefix:
            chart_name = f"charts/{chart_name}"
        if extension:
            chart_name = f"{chart_name}.html"

        return chart_name

    def __create_chart(
            self,
            ticker_id: str,
            from_date: datetime,
            to_date: datetime,
            indicators: bool,
    ) -> str:
        not_specified = ["", None]

        if ticker_id in not_specified:
            raise Exception(f"ticker_id must be specified")

        if from_date in not_specified or to_date in not_specified:
            raise Exception("from_date and to_date must be specified")

        normalized_df = self.__form_normalized_df(ticker_id)

        formed_name = ticker_id
        formed_title = _CHART_BASE_TITLE.format(
            ticker_id,
            from_date.strftime(_NORMALIZED_DATE_FORMAT),
            to_date.strftime(_NORMALIZED_DATE_FORMAT),
        )

        qf = cf.QuantFig(
            normalized_df,
            title=formed_title,
            name=formed_name,
        )
        if indicators:
            qf.add_volume()
            qf.add_rsi(color='black')
            qf.add_bollinger_bands(boll_std=2, colors=['grey'], fill=True)

        chart_path = qf.iplot(
            asUrl=True,
            filename=self.__form_chart_filename(
                ticker_id, from_date, to_date, indicators,
                prefix=True, extension=False,
            ),
        )
        return chart_path

    def create_chart(
            self,
            ticker_id: str,
            from_date: datetime,
            to_date: datetime,
            indicators: bool,
            force_refresh: bool,
    ) -> str:
        chart_filename = self.__form_chart_filename(
            ticker_id, from_date, to_date, indicators,
            prefix=True, extension=True,
        )
        if not force_refresh and path.exists(chart_filename) and path.isfile(chart_filename):
            return chart_filename

        chart_path = self.__create_chart(ticker_id, from_date, to_date, indicators)
        return chart_path






