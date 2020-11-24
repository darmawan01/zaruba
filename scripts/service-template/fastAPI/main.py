from fastapi import FastAPI

import os
import transport

# init app
app = FastAPI()

# init messagebus
mb: transport.MessageBus = transport.RMQMessageBus(
    rmq_host = os.getenv("SEED_RABBITMQ_HOST", "localhost"),
    rmq_user = os.getenv("SEED_RABBITMQ_USER", "root"),
    rmq_pass = os.getenv("SEED_RABBITMQ_PASS", "toor"),
    rmq_vhost = os.getenv("SEED_RABBITMQ_VHOST", "/")
)