package main

func execZFSMock(res string, err error) func(string, ...string) (string, error) {
	return func(first string, rest ...string) (string, error) {
		return res, err
	}
}
