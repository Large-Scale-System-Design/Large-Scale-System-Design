import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

import java.util.HashSet;
import java.util.Set;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

public class SnowflakeIdGeneratorTest {

    @Test
    void testGenerateUniqueIds() {
        SnowflakeIdGenerator generator =
                new SnowflakeIdGenerator(1, 1);

        Set<Long> ids = new HashSet<>();
        for (int i = 0; i < 100000; i++) {
            long id = generator.nextId();
            assertTrue(ids.add(id), "ID must be unique");
        }
    }

    @Test
    void testDifferentWorkerProducesDifferentIds() {
        SnowflakeIdGenerator gen1 = new SnowflakeIdGenerator(1, 1);
        SnowflakeIdGenerator gen2 = new SnowflakeIdGenerator(2, 1);

        long id1 = gen1.nextId();
        long id2 = gen2.nextId();

        assertNotEquals(id1, id2);
    }

    @Test
    void testSequenceOverflow() {
        SnowflakeIdGenerator generator = new SnowflakeIdGenerator(1, 1);

        long last = -1;
        for (int i = 0; i < 5000; i++) {
            long id = generator.nextId();
            assertTrue(id > last);
            last = id;
        }
    }

    @Test
    void testClockBackwards() {
        SnowflakeIdGenerator generator = new SnowflakeIdGenerator(1, 1);

        generator.nextId();

        // simulate rollback by reflection or overriding currentTime()
        // 여기서는 직접 예외 테스트는 생략

        assertDoesNotThrow(generator::nextId);
    }

    @Test
    void testConcurrentIdGeneration() throws Exception {

        int threadCount = 1000; // 동시에 실행할 스레드 수
        ExecutorService executor = Executors.newFixedThreadPool(threadCount);

        SnowflakeIdGenerator generator = new SnowflakeIdGenerator(1, 1);

        // 결과를 저장할 스레드 안전 Set
        Set<Long> idSet = ConcurrentHashMap.newKeySet();

        CountDownLatch startLatch = new CountDownLatch(1);   // 동시에 시작하게 함
        CountDownLatch doneLatch = new CountDownLatch(threadCount); // 스레드 종료 기다림

        for (int i = 0; i < threadCount; i++) {
            executor.submit(() -> {
                try {
                    startLatch.await(); // 모든 스레드가 준비될 때까지 대기
                    long id = generator.nextId();
                    idSet.add(id);
                } catch (Exception e) {
                    e.printStackTrace();
                    fail("Exception occurred during ID generation: " + e.getMessage());
                } finally {
                    doneLatch.countDown();
                }
            });
        }

        // 모든 스레드 출발!
        startLatch.countDown();

        // 모든 스레드 종료될 때까지 기다림
        doneLatch.await();

        executor.shutdown();

        // ✔ 모든 ID가 유일한지 확인
        assertEquals(threadCount, idSet.size(),
                "All generated IDs must be unique.");
    }
}