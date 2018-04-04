from enum import Enum


class Cmd(Enum):
    ADD = 0
    QUOTE = 1
    BUY = 2
    COMMIT_BUY = 3
    CANCEL_BUY = 4
    SELL = 5
    COMMIT_SELL = 6
    CANCEL_SELL = 7
    SET_BUY_AMOUNT = 8
    CANCEL_SET_BUY = 9
    SET_BUY_TRIGGER = 10
    SET_SELL_AMOUNT = 11
    SET_SELL_TRIGGER = 12
    CANCEL_SET_SELL = 13
    DUMPLOG = 14