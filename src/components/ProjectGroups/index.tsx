import { FC } from 'react'
import { groupEntries, Project } from '../../app/[project]/lib'
import { ProjectLink } from '../../app/[project]/components/ProjectLink'
import styles from './project-groups.module.css'

export const ProjectGroups: FC<{ projects: Project[] }> = ({ projects }) => {
  return (
    <div className={styles.groups}>
      {groupEntries(projects).map(([key, projects]) => (
        <section className={styles.group} key={key}>
          <header className={styles.header}>
            <h3>{key}</h3>
          </header>

          <div className={styles.list}>
            {projects!.map((project) => (
              <ProjectLink key={project.id} project={project} />
            ))}
          </div>
        </section>
      ))}
    </div>
  )
}
