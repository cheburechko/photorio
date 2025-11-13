def test_worker(kafka_task_writer, postgresql, worker):
    kafka_task_writer.send("tasks", b"test")
    kafka_task_writer.flush()
    assert False
