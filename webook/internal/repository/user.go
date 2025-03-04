package repository

type UserRepository struct {
}

func (r *UserRepository) FindById(int64) {
	// 先从cache找
	// 再从dao找
	// 找到写回cache

}
