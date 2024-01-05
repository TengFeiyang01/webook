package repository

type UserRepository struct {
}

func (r *UserRepository) FindById(int64) {
	// 先从 cache 找
	// 再从 dao 里面找
	// 找到了再往回写
}
