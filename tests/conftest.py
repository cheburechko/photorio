import os
import subprocess

from kafka import KafkaConsumer, KafkaProducer
from kafka.admin import KafkaAdminClient, NewTopic
import pytest
from testcontainers.elasticsearch import ElasticSearchContainer
from testcontainers.postgres import PostgresContainer
from testcontainers.kafka import KafkaContainer
import yaml


@pytest.fixture(scope="session")
def elasticsearch():
    with ElasticSearchContainer("library/elasticsearch:8.12.0") as container:
        host = container.get_container_host_ip()
        port = container.get_exposed_port(9200)

        yield host, port


@pytest.fixture(scope="session")
def postgresql():
    with PostgresContainer("postgres:16", driver=None) as container:
        yield container


@pytest.fixture(scope="session")
def kafka():
    with KafkaContainer() as container:
        yield [container.get_bootstrap_server()]


@pytest.fixture(scope="session")
def kafka_group():
    return "test"


@pytest.fixture(scope="session")
def kafka_admin_client(kafka):
    return KafkaAdminClient(bootstrap_servers=kafka)


@pytest.fixture(scope="function")
def kafka_task_topic(kafka_admin_client):
    topic = NewTopic("tasks", num_partitions=1, replication_factor=1)
    kafka_admin_client.create_topics([topic])
    yield topic.name
    kafka_admin_client.delete_topics([topic.name])


@pytest.fixture(scope="function")
def kafka_task_reader(kafka, kafka_task_topic):
    consumer = KafkaConsumer(kafka_task_topic, bootstrap_servers=kafka)
    yield consumer
    consumer.close()


@pytest.fixture(scope="function")
def kafka_task_writer(kafka, kafka_task_topic):
    producer = KafkaProducer(bootstrap_servers=kafka)
    yield producer
    producer.close()


def run_binary(args, test_output_dir):
    name = os.path.basename(args[0])
    with open(test_output_dir(f"{name}.err"), "w") as err:
        with open(test_output_dir(f"{name}.out"), "w") as out:
            process = subprocess.Popen(args, stdout=out, stderr=err)
            yield process
            process.terminate()
            try:
                process.wait(5)
            except subprocess.TimeoutExpired:
                process.kill()


@pytest.fixture(scope="session")
def root_dir():
    dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    yield lambda supath: os.path.join(dir, supath)


@pytest.fixture(scope="session")
def test_output_dir(root_dir):
    dir = root_dir(".test_output")
    os.makedirs(dir, exist_ok=True)
    yield lambda supath: os.path.join(dir, supath)


@pytest.fixture(scope="function")
def worker_config(postgresql, kafka, kafka_task_topic, elasticsearch, root_dir):
    return {
        "template_glob": root_dir("web/*"),
        "elasticsearch": {
            "Addresses": [elasticsearch],
        },
        "postgres": {
            "connection_url": postgresql.get_connection_url(),
        },
        "kafka": {
            "brokers": kafka,
            "task_topic": kafka_task_topic,
            "group": "worker",
            "poll_period": "60s",
        },
    }


@pytest.fixture(scope="function")
def worker_config_path(test_output_dir, worker_config):
    worker_config_path = test_output_dir("worker.yaml")
    with open(worker_config_path, "w") as worker_config_file:
        yaml.dump(worker_config, worker_config_file)
    yield worker_config_path


@pytest.fixture(scope="function")
def worker(root_dir, worker_config_path, test_output_dir):
    yield from run_binary(
        [root_dir("backend/cmd/worker/worker"), "--config", worker_config_path],
        test_output_dir,
    )
