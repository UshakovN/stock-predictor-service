import os
import datetime

import psycopg2
from psycopg2.extensions import connection

from dotenv import load_dotenv
from loguru import logger as log


_PG_DBNAME = "PG_DBNAME"
_PG_USER = "PG_USER"
_PG_PASSWORD = "PG_PASSWORD"
_PG_HOST = "PG_HOST"
_PG_PORT = "PG_PORT"


class ModelInfo:
    model_id: str
    current: bool
    accuracy: float
    created_at: datetime

    def __init__(
            self,
            model_id: str,
            accuracy: float,
            created_at: datetime,
    ):
        self.model_id = model_id
        self.accuracy = accuracy
        self.created_at = created_at


class Predict:
    predict_id: str
    ticker_id: str
    model_id: str
    date_predict: datetime
    predicted_movement: int
    predict_created_at: datetime

    def __init__(
            self,
            predict_id: str,
            ticker_id: str,
            model_id: str,
            date_predict: datetime,
            predicted_movement: int,
            predict_created_at: datetime,
    ):
        self.predict_id = predict_id
        self.ticker_id = ticker_id
        self.model_id = model_id
        self.date_predict = date_predict
        self.predicted_movement = predicted_movement
        self.predict_created_at = predict_created_at


class Config:
    dbname: str
    user: str
    password: str
    host: str
    port: int
    __loaded: bool

    def load(self, config_path: str):
        load_dotenv(config_path)

        dbname = os.getenv(_PG_DBNAME)
        user = os.getenv(_PG_USER)
        password = os.getenv(_PG_PASSWORD)
        host = os.getenv(_PG_HOST)
        port = os.getenv(_PG_PORT)

        for field in [dbname, user, password, host, port]:
            if field in [None, ""]:
                raise Exception(f"some required fields not found in config {config_path}")

        self.dbname = dbname
        self.user = user
        self.password = password
        self.host = host
        self.port = int(port)
        self.__loaded = True

    def loaded(self):
        return self.__loaded


class PredictsStorage:
    __conn: connection

    def __init__(self, config: Config):
        if not config.loaded():
            raise Exception("storage config has not loaded")

        self.__conn = psycopg2.connect(
            dbname=config.dbname,
            user=config.user,
            password=config.password,
            host=config.host,
            port=config.port,
        )
        try:
            self.__ping()
        except Exception as ex:
            raise Exception(f"cannot ping storage: {str(ex)}")

    def __ping(self):
        try:
            cur = self.__conn.cursor()
            cur.execute("SELECT 1")

            single_row = 1
            if len(cur.fetchone()) != single_row:
                raise Exception("ping select query must contains single row")

            self.__conn.commit()
            cur.close()
            log.info("storage ping query completed success")

        except Exception as ex:
            raise Exception(f"cannot execute ping select query: {str(ex)}")

    def put_model_info(self, model_info: ModelInfo):
        try:
            cur = self.__conn.cursor()
            cur.execute("""update model_info set current = %s""", [False])

            query = """
                insert into model_info (
                    model_id,
                    current,
                    accuracy,
                    created_at
                )
                values (%s, %s, %s, %s)
            """
            model_info.current = True
            args = [
                model_info.model_id,
                model_info.current,
                model_info.accuracy,
                model_info.created_at,
            ]
            cur.execute(query, args)
            self.__conn.commit()
            cur.close()

        except Exception as ex:
            raise Exception(f"cannot execute update or insert queries: {str(ex)}")

    def put_predicts(self, predicts: list[Predict]):
        query = """
            insert into stock_predict (
                predict_id,
                ticker_id,
                model_id,
                date_predict,
                predicted_movement,
                predict_created_at
            ) 
            values
        """
        args = []

        for predict in predicts:
            query = f"{query} (%s, %s, %s, %s, %s, %s), "
            args.extend([
                predict.predict_id,
                predict.ticker_id,
                predict.model_id,
                predict.date_predict,
                predict.predicted_movement,
                predict.predict_created_at,
            ])
        query = query.rstrip(", ")

        try:
            cur = self.__conn.cursor()
            cur.execute(query, args)
            self.__conn.commit()
            cur.close()

        except Exception as ex:
            raise Exception(f"cannot execute insert query: {str(ex)}")

















