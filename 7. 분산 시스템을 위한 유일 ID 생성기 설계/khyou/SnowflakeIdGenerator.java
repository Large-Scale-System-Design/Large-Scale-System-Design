/**
 * Twitter Snowflake ID Generator 구현체
 *
 * 64비트 ID 구성:
 * 0 - 41비트: timestamp (커스텀 epoch 이후 경과 밀리초)
 * 42 - 46비트: datacenterId (5 비트)
 * 47 - 51비트: workerId (5 비트)
 * 52 - 63비트: sequence (12 비트)
 */
public class SnowflakeIdGenerator {

    /** Snowflake의 기준 시각(Epoch). 2021-01-01 00:00:00 */
    private static final long EPOCH = 1609459200000L;

    /** workerId 비트 길이 = 5비트 (0~31까지 사용 가능) */
    private static final long WORKER_ID_BITS = 5L;

    /** datacenterId 비트 길이 = 5비트 (0~31까지 사용 가능) */
    private static final long DATACENTER_ID_BITS = 5L;

    /** 시퀀스 번호 비트 길이 = 12비트 (0~4095까지 사용 가능) */
    private static final long SEQUENCE_BITS = 12L;


    /** workerId 최댓값 계산: (2^5 - 1 = 31) */
    private static final long MAX_WORKER_ID = -1L ^ (-1L << WORKER_ID_BITS);

    /** datacenterId 최댓값 계산: (2^5 - 1 = 31) */
    private static final long MAX_DATACENTER_ID = -1L ^ (-1L << DATACENTER_ID_BITS);


    /** workerId는 sequence 뒤에 위치해야 하므로 sequence 비트 수만큼 shift */
    private static final long WORKER_ID_SHIFT = SEQUENCE_BITS;

    /** datacenterId는 workerId 뒤에 와야 하므로 sequence + workerId 만큼 shift */
    private static final long DATACENTER_ID_SHIFT = SEQUENCE_BITS + WORKER_ID_BITS;

    /** timestamp는 최상위에 오므로 sequence + workerId + datacenterId 만큼 shift */
    private static final long TIMESTAMP_LEFT_SHIFT =
            SEQUENCE_BITS + WORKER_ID_BITS + DATACENTER_ID_BITS;

    /** sequence mask: 12비트 최대값(4095)까지만 유지 */
    private static final long SEQUENCE_MASK = -1L ^ (-1L << SEQUENCE_BITS);


    /** 현재 인스턴스의 worker ID */
    private long workerId;

    /** 현재 인스턴스의 datacenter ID */
    private long datacenterId;

    /**
     * 동일한 밀리초 내에서 증가시키는 sequence 번호
     * 0 ~ 4095 순환
     */
    private long sequence = 0L;

    /** 마지막으로 ID를 생성한 timestamp 기록 */
    private long lastTimestamp = -1L;



    /**
     * 생성자: workerId & datacenterId 값을 받아 설정
     */
    public SnowflakeIdGenerator(long workerId, long datacenterId) {

        // workerId 범위 체크
        if (workerId > MAX_WORKER_ID || workerId < 0) {
            throw new IllegalArgumentException(String.format(
                    "worker Id can't be greater than %d or less than 0", MAX_WORKER_ID));
        }

        // datacenterId 범위 체크
        if (datacenterId > MAX_DATACENTER_ID || datacenterId < 0) {
            throw new IllegalArgumentException(String.format(
                    "datacenter Id can't be greater than %d or less than 0", MAX_DATACENTER_ID));
        }

        this.workerId = workerId;
        this.datacenterId = datacenterId;
    }



    /**
     * ID 생성 메서드 (동기화)
     * Snowflake ID는 멀티스레드 환경에서도 유일해야 하므로 synchronized 적용
     */
    public synchronized long nextId() {

        long timestamp = currentTime(); // 현재 밀리초 시각 읽기

        // 1. 시스템 시간이 과거로 돌아간 경우 → ID 충돌 가능성이 있으므로 거부
        if (timestamp < lastTimestamp) {
            throw new RuntimeException("Clock moved backwards. Refusing to generate id");
        }

        // 2. 같은 밀리초 안에서 여러 요청이 들어온 경우
        if (timestamp == lastTimestamp) {

            /**
             * sequence 증가:
             * - 같은 밀리초 내에서 0~4095까지 증가
             * - 4095 초과하면 mask 적용 → 0으로 돌아감
             */
            sequence = (sequence + 1) & SEQUENCE_MASK;

            // sequence가 0이 되었다면 → 같은 밀리초에서 4096개를 모두 사용한 것
            if (sequence == 0) {
                // 다음 밀리초가 될 때까지 대기해야 함
                timestamp = waitNextMillis(lastTimestamp);
            }

        } else {
            // 밀리초가 변경되면 sequence를 0으로 초기화 (새 밀리초이므로)
            sequence = 0L;
        }


        // 현재 timestamp를 lastTimestamp로 기록
        lastTimestamp = timestamp;

        /**
         * 3. 최종 Snowflake ID를 구성하는 부분
         *
         * timestamp 파트(41bit)  << 22
         * datacenterId(5bit)     << 17
         * workerId(5bit)         << 12
         * sequence(12bit)
         */
        return ((timestamp - EPOCH) << TIMESTAMP_LEFT_SHIFT)  // timestamp
                | (datacenterId << DATACENTER_ID_SHIFT)       // datacenterId
                | (workerId << WORKER_ID_SHIFT)               // workerId
                | sequence;                                   // sequence
    }



    /**
     * 동일한 millisecond 안에서 sequence overflow가 발생했을 때
     * 다음 millisecond까지 busy-wait(Spin-wait)
     */
    private long waitNextMillis(long lastTimestamp) {
        long timestamp = currentTime();

        // timestamp가 이전 밀리초보다 커질 때까지 루프
        while (timestamp <= lastTimestamp) {
            timestamp = currentTime();
        }

        return timestamp;
    }


    /** 현재 시간(밀리초)을 반환 */
    private long currentTime() {
        return System.currentTimeMillis();
    }
}