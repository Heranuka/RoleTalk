import logging

def setup_logger():
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - [%(request_id)s] %(message)s',
        datefmt='%Y-%m-%d %H:%M:%S'
    )
    old_factory = logging.getLogRecordFactory()
    def record_factory(*args, **kwargs):
        record = old_factory(*args, **kwargs)
        if not hasattr(record, 'request_id'):
            record.request_id = 'GLOBAL'
        return record
    logging.setLogRecordFactory(record_factory)
    return logging.getLogger("RoleTalk-AI")

logger = setup_logger()