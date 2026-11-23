// SPDX-License-Identifier: BSD-3-Clause
#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>
#include <utmp.h>

int main(void) {
    uint16_t endian = 1;
    const char *byte_order = *(unsigned char *)&endian ? "le" : "be";
    char filename[64];
    snprintf(filename, sizeof(filename), "utmp_%zu_%s", sizeof(struct utmp), byte_order);

    struct utmp records[3];
    memset(records, 0, sizeof(records));

    records[0].ut_type = BOOT_TIME;
    strcpy(records[0].ut_user, "ignored-boot");
    records[1].ut_type = USER_PROCESS;
    records[1].ut_pid = 1234;
    strcpy(records[1].ut_user, "audit-user");
    strcpy(records[1].ut_line, "pts/42");
    strcpy(records[1].ut_host, "example.test");
    strcpy(records[1].ut_id, "42");
    records[1].ut_session = 5678;
    records[1].ut_tv.tv_sec = 1700000000;
    records[1].ut_tv.tv_usec = 123456;
    records[2].ut_type = DEAD_PROCESS;
    strcpy(records[2].ut_user, "ignored-dead");

    FILE *file = fopen(filename, "wb");
    if (file == NULL) {
        perror("fopen");
        return 1;
    }
    if (fwrite(records, sizeof(records), 1, file) != 1) {
        perror("fwrite");
        fclose(file);
        return 1;
    }
    if (fclose(file) != 0) {
        perror("fclose");
        return 1;
    }

    printf("record=%zu byte_order=%s session=%zu timestamp=%zu tv_offset=%zu\n",
           sizeof(struct utmp), byte_order,
           sizeof(records[1].ut_session), sizeof(records[1].ut_tv),
           offsetof(struct utmp, ut_tv));
    return 0;
}
