#!/usr/local/bin/python3
# coding: utf-8

# status - setup.py
# 11/5/20 18:04
#
__author__ = "Benny <benny.think@gmail.com>"

from distutils.core import setup

# Package meta-data.
NAME = 'tgbot-ping'
DESCRIPTION = 'A package to detect telegram bot status in docker.'
URL = 'https://github.com/tgbot-collection/tgbot_ping'
EMAIL = 'benny.think@gmail.com'
AUTHOR = 'BennyThink'
REQUIRES_PYTHON = '>=3.6.0'
VERSION = "1.0.5"

# What packages are required for this module to be executed?
REQUIRED = ["requests"]

# Where the magic happens:
setup(
    name=NAME,
    version=VERSION,
    description=DESCRIPTION,
    author=AUTHOR,
    author_email=EMAIL,
    python_requires=REQUIRES_PYTHON,
    url=URL,
    packages=['tgbot_ping'],
    install_requires=REQUIRED,
    license='Apache 2.0',
    classifiers=[
        # Trove classifiers
        # Full list: https://pypi.python.org/pypi?%3Aaction=list_classifiers
        'License :: OSI Approved :: MIT License',
        'Programming Language :: Python',
        'Programming Language :: Python :: 3',
        'Programming Language :: Python :: 3.6',
        'Programming Language :: Python :: 3.7',
        'Programming Language :: Python :: 3.8',
        'Programming Language :: Python :: 3.9',
        'Programming Language :: Python :: Implementation :: CPython',
        'Programming Language :: Python :: Implementation :: PyPy'
    ],
    # data_files=['README.rst']
)
