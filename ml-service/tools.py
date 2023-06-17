import datetime
import hashlib
import uuid


def create_hash(message: str, salt: str) -> str:
    return hashlib.sha256(f"{message}{salt}".encode('utf-8')).hexdigest()


def create_uuid() -> str:
    try:
        created = uuid.uuid4()
    except Exception as ex:
        raise Exception(f"cannot create uuid: {ex}")
    return str(created)


def get_utc_time():
    return datetime.datetime.utcnow()
