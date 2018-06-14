/*
 *    This file is part of Disgo-Commons library.
 *
 *    The Disgo-Commons library is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *
 *    The Disgo-Commons library is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *
 *    You should have received a copy of the GNU General Public License
 *    along with the Disgo-Commons library.  If not, see <http://www.gnu.org/licenses/>.
 */
package types

//func init()  {
//	c = cache.New(CacheTTL, CacheTTL*2)
//	utils.Info("opening DB...")
//	opts := badger.DefaultOptions
//	opts.Dir = "." + string(os.PathSeparator) + "testdb"
//	opts.ValueDir = "." + string(os.PathSeparator) + "testdb"
//	db, _ = badger.Open(opts)
//}
//var testBlockByte = []byte("{\"id\":123,\"hash\":\"abc123\",\"numberOfTransactions\":1,\"updated\":\"2018-05-09T15:04:05Z\",\"created\":\"2018-05-09T15:04:05Z\"}")
//
//func TestBlockKey(t *testing.T) {
//	block := &Block{}
//	block.Id = int64(123)
//	if block.Key() != "block-123" {
//		t.Errorf("block.Key() returning invalid value: %s", block.Key())
//	}
//}
//
//func TestBlockUnmarshalJSON(t *testing.T) {
//	block := &Block{}
//	block.UnmarshalJSON(testBlockByte)
//	if block.Id != int64(123) {
//		t.Errorf("block.UnmarshalJSON returning invalid %s value: %d", "Id", block.Id)
//	}
//	if block.Hash != "abc123" {
//		t.Errorf("block.UnmarshalJSON returning invalid %s value: %s", "Hash", block.Hash)
//	}
//	if block.NumberOfTransactions != int64(1) {
//		t.Errorf("block.UnmarshalJSON returning invalid %s value: %d", "NumberOfTransactions", block.NumberOfTransactions)
//	}
//	d, _ := time.Parse(time.RFC3339, "2018-05-09T15:04:05Z")
//	if block.Updated != d {
//		t.Errorf("block.UnmarshalJSON returning invalid %s value: %s", "Updated", block.Updated.String())
//	}
//	if block.Created != d {
//		t.Errorf("block.UnmarshalJSON returning invalid %s value: %s", "Created", block.Created.String())
//	}
//}
//
//func TestBlockMarshalJSON(t *testing.T) {
//	block := &Block{}
//	block.UnmarshalJSON(testBlockByte)
//	out, err := block.MarshalJSON()
//	if err != nil {
//		t.Fatalf("block.MarshalJSON returning error: %s", err)
//	}
//	if reflect.DeepEqual(out, testBlockByte) == false {
//		t.Errorf("block.MarshalJSON returning invalid value.\nGot: %s\nExpected: %s", out, testBlockByte)
//	}
//}
