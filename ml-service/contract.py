from pydantic import BaseModel
from typing import Union
from datetime import datetime


class ChartRequest(BaseModel):
    from_date: datetime
    to_date: datetime
    indicators: Union[bool, None] = None
    force_refresh: Union[bool, None] = None


class ChartResponse(BaseModel):
    success: bool
    chart_url: str


class HealthResponse(BaseModel):
    success: bool

