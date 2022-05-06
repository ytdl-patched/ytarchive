import ctypes

ytarchive = ctypes.CDLL("./ytarchive.so")
ytarchive.initialize.argtypes = [ctypes.c_char_p]
ytarchive.initialize.restype = ctypes.c_long
print(ytarchive.initialize(b"11111111111"))
