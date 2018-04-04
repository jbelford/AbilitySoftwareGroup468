class Command(object):
	C_type = None
	TransactionID = None
	UserID = None
	Amount = None
	StockSymbol = None
	FileName = None
	TimeStamp = None

	def __init__(self, C_type=None,
	             TransactionID=None,
	             UserID=None,
	             Amount=None,
	             StockSymbol=None,
	             FileName=None,
	             TimeStamp=None):
		self.C_type = C_type
		self.TransactionID = TransactionID
		self.UserID = UserID
		self.Amount = Amount
		self.StockSymbol = StockSymbol
		self.FileName = FileName
		self.TimeStamp = TimeStamp