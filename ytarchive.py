import ctypes
import json

from typing import Union

dynload = ctypes.CDLL("./_ytarchive.so")
dynload.goInfo.argtypes = []
dynload.goInfo.restype = ctypes.c_char_p
up_siz_str, = dynload.goInfo().decode().split(" ")
up_size = int(up_siz_str)

# find a corresponding type for uintptr in ctypes
for tp in (ctypes.c_uint, ctypes.c_ulong, ctypes.c_ulonglong):
    if ctypes.sizeof(tp) == up_size:
        go_uintptr = tp
        break
else:
    import warnings
    warnings.warn(ResourceWarning("Failed to figure out corresponding type for uintptr in ctypes. Assuming unsigned long long."))
    go_uintptr = ctypes.c_ulonglong

dynload.initialize.argtypes = [ctypes.c_char_p]
dynload.initialize.restype = go_uintptr

dynload.release.argtypes = [ctypes.c_char_p]
dynload.release.restype = None

dynload.registerFormat.argtypes = [go_uintptr, ctypes.c_char_p, ctypes.c_char_p, ctypes.c_char_p, ctypes.c_char_p]
dynload.registerFormat.restype = None

dynload.loadCookies.argtypes = [ctypes.c_char_p]
# golang has no corresponding type to _Bool. the returned value can be used in if statements without conversion
dynload.loadCookies.restype = ctypes.c_int

dynload.runDownloader.argtypes = [go_uintptr]
dynload.runDownloader.restype = None

dynload.interrupt.argtypes = [go_uintptr]
dynload.interrupt.restype = None

# timeout argument (which is the 2nd argument) is in milliseconds, unlike time.sleep
dynload.poll.argtypes = [go_uintptr, ctypes.c_int]
dynload.poll.restype = ctypes.c_char_p

sob = Union[str, bytes]


def to_bytes(*args: sob) -> list[bytes]:
    return [x.encode() if isinstance(x, str) else x for x in args]


def initialize(video_id: Union[str, bytes]) -> int:
    """
    Initialized a new downloader instance.
    The returned value is the handle to the golang object.
    """
    video_id = to_bytes(video_id)[0]
    return dynload.initialize(video_id)


def release(ptr: int):
    """
    Releases given handle from internal reference.
    The object will be gargage-collected in golang side.
    Calling this method multiple times against same handle or some random value is safe.
    DO NOT USE RELEASED HANDLE IN OTHER FUNCTIONS
    """
    dynload.release(ptr)


def register_format(
        ptr: int,
        format_id: sob, fmt_url: sob, manifest_url: sob, filepath: sob):
    """
    Registers a format for download. Call many times as requested.
    All argument names comply to yt-dlp's info_dict keys, except ptr.
    """
    format_id, fmt_url, manifest_url, filepath = to_bytes(format_id, fmt_url, manifest_url, filepath)
    dynload.registerFormat(ptr, format_id, fmt_url, manifest_url, filepath)


def load_cookies(cookie_file: sob) -> bool:
    """
    Load cookies from a file. The file must be in Netscape format.
    Loaded results are shared across instances.
    Returns whether loading is successful or not.
    """
    cookie_file = to_bytes(cookie_file)[0]
    return bool(dynload.loadCookies(cookie_file))


def run_downloader(ptr: int):
    """
    Runs download task in calling thread.
    It is highly recommended to call this method in background thread, e.g. threading.Thread.
    You can check status of the downloading by using poll().
    """
    dynload.runDownloader(ptr)


def interrupt(ptr: int):
    """
    Interrupt and stops download process.
    This has the same effect as hitting Ctrl+C.
    """
    dynload.interrupt(ptr)


def poll(ptr: int, timeout: int) -> dict:
    """
    Polling method for downloader, to check progress of download.
    JSON returned by golang side is decoded and returned.
    """
    # json.loads accepts both str and bytes
    return json.loads(dynload.poll(ptr, timeout))
