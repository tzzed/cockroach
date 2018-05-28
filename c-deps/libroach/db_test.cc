// Copyright 2017 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.  See the License for the specific language governing
// permissions and limitations under the License.

#include "db.h"
#include "include/libroach.h"
#include "status.h"
#include "testutils.h"

using namespace cockroach;

TEST(Libroach, DBOpenHook) {
  DBOptions db_opts;

  // Try an empty extra_options.
  db_opts.extra_options = ToDBSlice("");
  EXPECT_OK(DBOpenHook("", db_opts, nullptr));

  // Try extra_options with anything at all.
  db_opts.extra_options = ToDBSlice("blah");
  EXPECT_ERR(DBOpenHook("", db_opts, nullptr),
             "DBOptions has extra_options, but OSS code cannot handle them");
}

TEST(Libroach, DBOpen) {
  DBOptions db_opts = defaultDBOptions();
  DBEngine* db;

  EXPECT_STREQ(DBOpen(&db, DBSlice(), db_opts).data, NULL);
  DBEnvStatsResult stats;
  EXPECT_STREQ(DBGetEnvStats(db, &stats).data, NULL);
  EXPECT_STREQ(stats.encryption_status.data, NULL);

  DBClose(db);
}

TEST(Libroach, BatchSSTablesForCompaction) {
  auto toString = [](const std::vector<rocksdb::Range>& ranges) -> std::string {
    std::string res;
    for (auto r : ranges) {
      if (!res.empty()) {
        res.append(",");
      }
      res.append(r.start.data(), r.start.size());
      res.append("-");
      res.append(r.limit.data(), r.limit.size());
    }
    return res;
  };

  auto sst = [](const std::string& smallest, const std::string& largest,
                uint64_t size) -> rocksdb::SstFileMetaData {
    return rocksdb::SstFileMetaData("", "", size, 0, 0, smallest, largest, 0, 0);
  };

  struct TestCase {
    TestCase(const std::vector<rocksdb::SstFileMetaData>& s,
             const std::string& start, const std::string& end,
             uint64_t target, const std::string& expected)
        : sst(s),
          start_key(start),
          end_key(end),
          target_size(target),
          expected_ranges(expected) {
    }
    std::vector<rocksdb::SstFileMetaData> sst;
    std::string start_key;
    std::string end_key;
    uint64_t target_size;
    std::string expected_ranges;
  };

  std::vector<TestCase> testCases = {
    TestCase({ sst("a", "b", 10) },
             "", "", 10, "-"),
    TestCase({ sst("a", "b", 10) },
             "a", "", 10, "a-"),
    TestCase({ sst("a", "b", 10) },
             "", "b", 10, "-b"),
    TestCase({ sst("a", "b", 10) },
             "a", "b", 10, "a-b"),
    TestCase({ sst("c", "d", 10) },
             "a", "b", 10, "a-b"),
    TestCase({ sst("a", "b", 10), sst("b", "c", 10) },
             "a", "c", 10, "a-b,b-c"),
    TestCase({ sst("a", "b", 10), sst("b", "c", 10) },
             "a", "c", 100, "a-c"),
    TestCase({ sst("a", "b", 10), sst("b", "c", 10) },
             "", "c", 10, "-b,b-c"),
    TestCase({ sst("a", "b", 10), sst("b", "c", 10) },
             "a", "", 10, "a-b,b-"),
    TestCase({ sst("a", "b", 10), sst("b", "c", 10), sst("c", "d", 10) },
             "a", "d", 10, "a-b,b-c,c-d"),
    TestCase({ sst("a", "b", 10), sst("b", "c", 10), sst("c", "d", 10) },
             "a", "d", 20, "a-c,c-d"),
  };
  for (auto c : testCases) {
    std::vector<rocksdb::Range> ranges;
    BatchSSTablesForCompaction(c.sst, c.start_key, c.end_key, c.target_size, &ranges);
    auto result = toString(ranges);
    EXPECT_EQ(c.expected_ranges, result);
  }
}
