FROM python:3.10.7

ENV PYTHONDONTWRITEBYTECODE 1
ENV PYTHONUNBUFFERED 1

WORKDIR /code

RUN pip install --upgrade pip
RUN pip install pipenv

COPY Pipfile Pipfile.lock /code/
RUN pipenv install --dev --system

COPY . /code/
