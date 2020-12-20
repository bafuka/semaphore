package sql

import (
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/masterminds/squirrel"
	"time"
)

func (d *SqlDb) CreateProject(project db.Project) (newProject db.Project, err error) {
	project.Created = time.Now()

	res, err := d.sql.Exec("insert into project(name, created) values (?, ?)", project.Name, project.Created)
	if err != nil {
		return
	}

	insertId, err := res.LastInsertId()
	if err != nil {
		return
	}

	newProject = project
	newProject.ID = int(insertId)
	return
}

func (d *SqlDb) GetProjects(userID int) (projects []db.Project, err error) {
	query, args, err := squirrel.Select("p.*").
		From("project as p").
		Join("project__user as pu on pu.project_id=p.id").
		Where("pu.user_id=?", userID).
		OrderBy("p.name").
		ToSql()

	if err != nil {
		return
	}

	_, err = d.sql.Select(&projects, query, args...)

	return
}

func (d *SqlDb) GetProject(projectID int) (project db.Project, err error) {
	query, args, err := squirrel.Select("p.*").
		From("project as p").
		Where("p.id=?", projectID).
		ToSql()

	if err != nil {
		return
	}

	err = d.sql.SelectOne(&project, query, args...)

	return
}

func (d *SqlDb) DeleteProject(projectID int) error {
	tx, err := d.sql.Begin()

	if err != nil {
		return err
	}

	statements := []string{
		"delete from project__template where project_id=?",
		"delete from project__user where project_id=?",
		"delete from project__repository where project_id=?",
		"delete from project__inventory where project_id=?",
		"delete from access_key where project_id=?",
		"delete from project where id=?",
	}

	for _, statement := range statements {
		_, err = tx.Exec(statement, projectID)

		if err != nil {
			err = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (d *SqlDb) UpdateProject(project db.Project) error {
	_, err := d.sql.Exec(
		"update project set name=?, alert=?, alert_chat=? where id=?",
		project.Name,
		project.Alert,
		project.AlertChat,
		project.ID)
	return err
}
