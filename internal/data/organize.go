package data

import (
	"NakedVPN/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type OrganizeRepo struct {
	data *Data
	log  *log.Helper
}

func NewOrganizeRepo(data *Data, logger log.Logger) biz.OrganizeRepo {
	return &OrganizeRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "data/organize")),
	}
}

func (r *OrganizeRepo) GetAllOrganizes() ([]*biz.Organize, error) {
	var orgs []*biz.Organize
	if err := r.data.db.Find(&orgs).Error; err != nil {
		return nil, err
	}
	return orgs, nil
}
