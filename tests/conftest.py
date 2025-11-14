import contextlib
import os
import subprocess

from kafka import KafkaConsumer, KafkaProducer
from kafka.admin import KafkaAdminClient, NewTopic
import pytest
from testcontainers.elasticsearch import ElasticSearchContainer
from testcontainers.postgres import PostgresContainer
from testcontainers.kafka import KafkaContainer
from testcontainers.core.container import DockerContainer
from testcontainers.core.waiting_utils import wait_for_logs
from testcontainers.core.network import Network


@contextlib.contextmanager
def dump_logs(container, test_output_dir, name):
    with container:
        try:
            yield
        finally:
            logs = container.get_logs()
            with open(test_output_dir(f"{name}.out"), "w") as f:
                f.write(logs[0].decode("utf-8"))
            with open(test_output_dir(f"{name}.err"), "w") as f:
                f.write(logs[1].decode("utf-8"))


@pytest.fixture(scope="session")
def network():
    with Network() as network:
        yield network


@pytest.fixture(scope="session")
def elasticsearch(network, test_output_dir):
    container = (
        ElasticSearchContainer("library/elasticsearch:8.12.0")
        .with_network(network)
        .with_network_aliases("elasticsearch")
    )
    with dump_logs(container, test_output_dir, "elasticsearch"):
        yield container


@pytest.fixture(scope="session")
def postgresql(root_dir, test_output_dir, network):
    container = (
        PostgresContainer("postgres:16", driver=None)
        .with_network(network)
        .with_network_aliases("postgresql")
    )
    with dump_logs(container, test_output_dir, "postgresql"):
        connection_url = container.get_connection_url()
        os.environ["DATABASE_URL"] = connection_url
        os.environ["ROOT_DATABASE_URL"] = connection_url

        run_binary(
            [
                "graphile-migrate",
                "reset",
                "--config",
                root_dir("graphile/.gmrc"),
                "--erase",
            ],
            test_output_dir,
        )
        yield container


@pytest.fixture(scope="session")
def postgresql_alias(postgresql):
    return f"postgresql://{postgresql.username}:{postgresql.password}@postgresql:{postgresql.port}/{postgresql.dbname}"


@pytest.fixture(scope="session")
def kafka(network, test_output_dir):
    container = KafkaContainer().with_network(network).with_network_aliases("kafka")
    with dump_logs(container, test_output_dir, "kafka"):
        yield container


@pytest.fixture(scope="session")
def kafka_group():
    return "test"


@pytest.fixture(scope="session")
def kafka_alias(kafka):
    return f"kafka:{kafka.port}"


@pytest.fixture(scope="session")
def kafka_admin_client(kafka):
    return KafkaAdminClient(bootstrap_servers=kafka.get_bootstrap_server())


@pytest.fixture(scope="function")
def kafka_task_topic(kafka_admin_client):
    topic = NewTopic("tasks", num_partitions=1, replication_factor=1)
    kafka_admin_client.create_topics([topic])
    yield topic.name
    kafka_admin_client.delete_topics([topic.name])


@pytest.fixture(scope="function")
def kafka_task_reader(kafka, kafka_task_topic):
    consumer = KafkaConsumer(
        kafka_task_topic, bootstrap_servers=kafka.get_bootstrap_server()
    )
    yield consumer
    consumer.close()


@pytest.fixture(scope="function")
def kafka_task_writer(kafka):
    producer = KafkaProducer(bootstrap_servers=kafka.get_bootstrap_server())
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
def worker(postgresql_alias, kafka_alias, kafka_task_topic, test_output_dir, network):
    container = (
        DockerContainer("photorio/worker:latest")
        .with_network(network)
        .with_envs(
            POSTGRES_CONNECTION_URL=postgresql_alias,
            KAFKA_BROKERS=kafka_alias,
            TASK_TOPIC=kafka_task_topic,
        )
        .with_network_aliases("worker")
    )
    with dump_logs(container, test_output_dir, "worker"):
        wait_for_logs(container, "Starting consumer", timeout=10)
        yield container
