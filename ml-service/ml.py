import storage
import tools

import datetime
import logging as log
import numpy as np
import pandas as pd

from sklearn.ensemble import RandomForestClassifier
from sklearn.model_selection import train_test_split
from sklearn.metrics import accuracy_score


_DATE_FORMAT = '%Y-%m-%d'


class PriceMovementPredictor:
    __rfc: RandomForestClassifier

    __rfc_id: str
    __rfc_accuracy: float
    __rfc_build_dt: datetime

    __storage: storage.PredictsStorage
    __fixed_df: pd.DataFrame

    def __init__(self, config_path: str):
        storage_config = storage.Config()
        try:
            storage_config.load(config_path)
        except Exception as ex:
            raise Exception(f"predicts storage config loading failed: {str(ex)}")
        try:
            self.__storage = storage.PredictsStorage(storage_config)
        except Exception as ex:
            raise Exception(f"cannot create predicts storage: {str(ex)}")

    def __normalize_stocks(self):
        # strip time from stocked_time
        self.__fixed_df['stocked_time'] = pd.to_datetime(
            self.__fixed_df['stocked_time'],
        )
        self.__fixed_df['stocked_time'] = self.__fixed_df['stocked_time'].dt.strftime(_DATE_FORMAT)

        # sort stocks dataframe
        self.__fixed_df.sort_values(
            by=[
                'ticker_id',
                'stocked_time',
            ],
            inplace=True,
        )
        # smooth stocks price fields
        self.__exponential_smooth_stocks()

        # calculate price diff for each period
        self.__fixed_df['price_change'] = self.__fixed_df['close_price'].diff()

        # create mask for ticker_id shifting
        mask = self.__fixed_df['ticker_id'] != self.__fixed_df['ticker_id'].shift(1)

        # normalize price diff for different tickers
        self.__fixed_df['price_change'] = np.where(
            mask,
            np.nan,
            self.__fixed_df['price_change'],
        )

    def __exponential_smooth_stocks(self, alpha_factor: int = 0.0095):
        smoothed_df = self.__fixed_df.groupby('ticker_id')[[
            'open_price',
            'close_price',
            'highest_price',
            'lowest_price',
            'trading_volume',
        ]].transform(
            lambda x: x.ewm(alpha=alpha_factor).mean()
        )
        smoothed_df = pd.concat([
            self.__fixed_df[['ticker_id', 'stocked_time']],
            smoothed_df,
        ],
            axis=1,
            sort=False,
        )
        self.__fixed_df = smoothed_df

    def __fill_rsi_indicator(self, period: int = 27):
        sample_df = self.__fixed_df[['ticker_id', 'price_change']].copy()

        up_df = sample_df.copy()
        up_df.loc['price_change'] = up_df.loc[(up_df['price_change'] < 0), 'price_change'] = 0

        down_df = sample_df.copy()
        down_df.loc['price_change'] = down_df.loc[(down_df['price_change'] > 0), 'price_change'] = 0

        down_df['price_change'] = down_df['price_change'].abs()

        def form_moving_window(df: pd.DataFrame):
            return df.groupby('ticker_id')['price_change'].transform(
                lambda x: x.ewm(span=period).mean(),
            )

        ewm_up = form_moving_window(up_df)
        ewm_down = form_moving_window(down_df)

        relative_strength = ewm_up / ewm_down
        relative_strength_index = 100.0 - (100.0 / (1.0 + relative_strength))

        self.__fixed_df['rsi_indicator'] = relative_strength_index

    def __calculate_min_for_lowest(self, period: int = 14) -> pd.DataFrame:
        lowest_df = self.__fixed_df[['ticker_id', 'lowest_price']].copy()

        min_for_lowest = lowest_df.groupby('ticker_id')['lowest_price'].transform(
            lambda x: x.rolling(window=period).min()
        )
        return min_for_lowest

    def __calculate_max_for_highest(self,  period: int = 14) -> pd.DataFrame:
        highest_df = self.__fixed_df[['ticker_id', 'highest_price']].copy()

        max_for_highest = highest_df.groupby('ticker_id')['highest_price'].transform(
            lambda x: x.rolling(window=period).max()
        )
        return max_for_highest

    def __fill_stochastic_indicator(self, period: int = 14):
        lowest_df = self.__calculate_min_for_lowest(period)
        highest_df = self.__calculate_max_for_highest(period)

        close_df = self.__fixed_df['close_price']
        stochastic = 100.0 * ((close_df - lowest_df) / (highest_df - lowest_df))

        self.__fixed_df['stochastic_indicator'] = stochastic

    def __fill_williams_percent_range_indicator(self, period: int = 14):
        lowest_df = self.__calculate_min_for_lowest(period)
        highest_df = self.__calculate_max_for_highest(period)

        close_df = self.__fixed_df['close_price']
        williams_pr = -100.0 * ((highest_df - close_df) / (highest_df - lowest_df))

        self.__fixed_df['williams_indicator'] = williams_pr

    def __fill_roc_indicator(self, period: int = 21):
        roc = self.__fixed_df.groupby('ticker_id')['close_price'].transform(
            lambda x: x.pct_change(periods=period)
        )
        self.__fixed_df['roc_indicator'] = roc

    def __fill_macd_sl_indicator(self):
        def form_ema(df: pd.DataFrame, period: int) -> pd.DataFrame:
            return df.groupby('ticker_id')['close_price'].transform(
                lambda x: x.ewm(span=period).mean()
            )

        ema_26 = form_ema(self.__fixed_df, period=26)
        ema_12 = form_ema(self.__fixed_df, period=12)

        macd: pd.DataFrame = ema_12 - ema_26
        signal_line = macd.ewm(span=9).mean()

        self.__fixed_df['macd_indicator'] = macd
        self.__fixed_df['sl_indicator'] = signal_line

    def __fill_classification_factor(self):
        close_df = self.__fixed_df.groupby('ticker_id')['close_price'].transform(
            # shift for correct predict column
            lambda x: np.sign(x.diff()).shift(1)
        )
        self.__fixed_df['prediction'] = close_df

        # set positive predict for non movement price
        self.__fixed_df.loc[self.__fixed_df['prediction'] == 0.0] = 1.0

    def __sanitize_stocks(self):
        self.__fixed_df = self.__fixed_df.dropna()

    def __build_stocks_dataset(self) -> pd.DataFrame:
        self.__normalize_stocks()

        self.__fill_rsi_indicator(period=27)
        self.__fill_stochastic_indicator(period=14)
        self.__fill_williams_percent_range_indicator(period=14)
        self.__fill_roc_indicator(period=21)
        self.__fill_macd_sl_indicator()
        self.__fill_classification_factor()

        self.__sanitize_stocks()

        return self.__fixed_df

    def __build_past_date_data(self, days_delta: int = 1) -> pd.DataFrame:
        past_date = tools.get_utc_time().now() - datetime.timedelta(days=days_delta)
        past_date_format = past_date.strftime(_DATE_FORMAT)

        dataset = self.__fixed_df
        past_date_data = dataset.query(f"stocked_time == '{past_date_format}'")

        return past_date_data

    @staticmethod
    def __fiter_dataset_indicators(dataset: pd.DataFrame) -> pd.DataFrame:
        return dataset[[
            'rsi_indicator',
            'stochastic_indicator',
            'williams_indicator',
            'roc_indicator',
            'macd_indicator',
            'sl_indicator',
        ]]

    @staticmethod
    def __split_stocks_dataset(dataset: pd.DataFrame):
        x_samples = PriceMovementPredictor.__fiter_dataset_indicators(dataset)
        y_samples = dataset['prediction']

        x_train, x_test, y_train, y_test = train_test_split(
            x_samples,
            y_samples,
            random_state=1,
            shuffle=False,
            stratify=None,
        )
        return x_train, x_test, y_train, y_test

    def __build_classifier(self):
        try:
            rfc = RandomForestClassifier(
                criterion="gini",
                n_estimators=100,
                random_state=1,
                oob_score=True,
            )
            log.info("random forest classifier model create")

        except Exception as ex:
            raise Exception(f"cannot create random forest classifier: {str(ex)}")

        try:
            dataset = self.__build_stocks_dataset()
            log.info("dataset for classifier model build success")

        except Exception as ex:
            raise Exception(f"cannot build dataset for classifier: {str(ex)}")

        try:
            x_train, x_test, y_train, y_test = self.__split_stocks_dataset(dataset)
            log.info("classifier dataset split success")

        except Exception as ex:
            raise Exception(f"cannot split dataset for classified fit: {str(ex)}")

        try:
            rfc.fit(x_train, y_train)
            log.info("random forest classifier model fit success")

        except Exception as ex:
            raise Exception(f"cannot train classifier to dataset: {str(ex)}")

        self.__rfc = rfc
        self.__rfc_id = tools.create_uuid()
        self.__rfc_build_dt = tools.get_utc_time()
        self.__rfc_accuracy = accuracy_score(y_test, self.__rfc.predict(x_test), normalize=True)

        model_info = storage.ModelInfo(
            model_id=self.__rfc_id,
            accuracy=self.__rfc_accuracy,
            created_at=self.__rfc_build_dt,
        )
        self.__storage.put_model_info(model_info)

        log.info("random forest classifier model build and fit success")

    def __build_price_movement_predict(self):
        monday_weekday: int = 0
        today_date = tools.get_utc_time().date()

        if today_date.weekday() != monday_weekday:
            days_backoff = 1
        else:
            days_backoff = 3

        past_date_data = self.__build_past_date_data(days_delta=days_backoff)

        if past_date_data.empty:
            log.warning("past date data not found for build price movement predict")
            return

        past_date_indicators = PriceMovementPredictor.__fiter_dataset_indicators(past_date_data)
        today_price_movement_predictions = self.__rfc.predict(past_date_indicators)

        ticker_ids = past_date_data['ticker_id'].to_list()

        ticker_ids_count = len(ticker_ids)
        predictions_count = len(today_price_movement_predictions)

        if ticker_ids_count != predictions_count:
            raise Exception(
                f"ticker_id count: {ticker_ids_count} not match to "
                f"price movement predictions count: {predictions_count}",
            )

        if any([predict not in [-1, 1] for predict in today_price_movement_predictions]):
            raise Exception("non binary values encountered in model predictions")

        today_date = tools.get_utc_time().date()
        predicts: list[storage.Predict] = []

        for idx in range(ticker_ids_count):
            predict = storage.Predict(
                predict_id=tools.create_uuid(),
                ticker_id=ticker_ids[idx],
                model_id=self.__rfc_id,
                date_predict=today_date,
                predicted_movement=int(today_price_movement_predictions[idx]),
                predict_created_at=tools.get_utc_time(),
            )
            predicts.append(predict)

        try:
            self.__storage.put_predicts(predicts)
        except Exception as ex:
            raise Exception(f"cannot put model price movement predicts to storage: {str(ex)}")

    def __cleanup_dataset(self):
        self.__fixed_df = pd.DataFrame(None)

    def update_stored_predicts(self, df: pd.DataFrame):
        self.__fixed_df = df
        try:
            self.__build_classifier()
        except Exception as ex:
            raise Exception(f"cannot build classifier: {str(ex)}")
        try:
            self.__build_price_movement_predict()
        except Exception as ex:
            raise Exception(f"cannot build price movement predict: {str(ex)}")

        self.__cleanup_dataset()
        log.info("cleanup dataset for reduce memory usage")


