#!/usr/local/bin/python3
# coding: utf-8

# tgbot_ping - __init__.py.py
# 11/5/20 18:13
#

__author__ = "Benny <benny.think@gmail.com>"

import datetime
import json
import os
import traceback

import requests


def get_runtime(container_name: "str", display_name: "str" = "This bot", parse_mode: "str" = "markdown",
                raw=False):
    try:
        info, raw_data = __get_container_info(container_name, display_name)
    except Exception:
        raw_data = ""
        info = f"Runtime information is not available outside of docker container.\n😊{traceback.format_exc()}😭"

    if parse_mode == "markdown":
        info = info.replace("😊", "`").replace("😭", "`")
    elif parse_mode == "html":
        info = info.replace("😊", "<pre>").replace("😭", "</pre>")
    else:
        raise ValueError(f"mode {parse_mode} is invalid.")
    if raw:
        return info, raw_data
    return info


def __get_container_info(container_name: str, display_name):
    # http://socat:2375/containers/untitled_socat_1/json
    # http://socat:2375/containers/osstpmgt_websvc_1/stats?stream=0
    msg_template = "{bot} has been running for 😊{run}😭 from " \
                   "😊{started_at}😭😄\n" \
                   "CPU: 😊{cpu}😭\n" \
                   "RAM: 😊{ram}😭\n" \
                   "Network RX/TX: 😊{rx}/{tx}😭\n" \
                   "IO R/W: 😊{R}/{W}😭\n"

    if os.getenv("PYTHON_ENV") == "dev":
        with open("stats.json", "r") as f:
            stats = json.load(f)
        with open("inspect.json", "r") as f:
            inspect = json.load(f)
    else:
        stats = requests.get(f"http://socat:2375/containers/{container_name}/stats?stream=0").json()
        inspect = requests.get(f"http://socat:2375/containers/{container_name}/json").json()

    start_time = inspect["State"]["StartedAt"][0:26]
    utc_time = datetime.datetime.strptime(start_time, "%Y-%m-%dT%H:%M:%S.%f")
    delta = datetime.timedelta(hours=8)
    run = datetime.datetime.now() - utc_time - delta
    localtime: str = (utc_time + delta).astimezone().strftime("%Y-%m-%d %H:%M:%S %z")

    io_stats = stats["blkio_stats"]["io_service_bytes_recursive"]
    io_read, io_write = "0B", "0B"
    if io_stats:
        for item in io_stats:
            if item["op"] == "Read":
                io_read = __human_bytes(item["value"])
            elif item["op"] == "Write":
                io_write = __human_bytes(item["value"])
    cpu = __calculate_cpu_percent(stats)
    ram = __human_bytes(stats["memory_stats"]["usage"])
    rx = __human_bytes(stats["networks"]["eth0"]["rx_bytes"])
    tx = __human_bytes(stats["networks"]["eth0"]["tx_bytes"])
    msg = msg_template.format(bot=display_name, run=str(run).split(".")[0],
                              started_at=localtime, cpu=cpu, ram=ram,
                              rx=rx, tx=tx, R=io_read, W=io_write)
    return msg, stats


def __calculate_cpu_percent(d: dict) -> str:
    # https://github.com/moby/moby/blob/eb131c5383db8cac633919f82abad86c99bffbe5/cli/command/container/stats_helpers.go#L175-L188
    cpu_count = d["cpu_stats"].get("online_cpus", 1)
    cpu_percent = 0.0
    cpu_delta = float(d["cpu_stats"]["cpu_usage"]["total_usage"] - d["precpu_stats"]["cpu_usage"]["total_usage"])
    system_delta = float(d["cpu_stats"].get("system_cpu_usage", 0) - d["precpu_stats"].get("system_cpu_usage", 0))
    if system_delta > 0.0:
        cpu_percent = cpu_delta / system_delta * 100.0 * cpu_count
    return "%.2f%%" % cpu_percent


def __human_bytes(byte: int) -> str:
    byte = float(byte)
    kb = float(1024)
    mb = float(kb ** 2)  # 1,048,576
    gb = float(kb ** 3)  # 1,073,741,824
    tb = float(kb ** 4)  # 1,099,511,627,776

    if byte < kb:
        return '{0}{1}'.format(byte, 'Bytes' if 0 == byte > 1 else 'B')
    elif kb <= byte < mb:
        return '{0:.2f}KB'.format(byte / kb)
    elif mb <= byte < gb:
        return '{0:.2f}MB'.format(byte / mb)
    elif gb <= byte < tb:
        return '{0:.2f}GB'.format(byte / gb)
    elif tb <= byte:
        return '{0:.2f}TB'.format(byte / tb)


if __name__ == '__main__':
    x = get_runtime("vigilant_antonelli")
    print(x)
