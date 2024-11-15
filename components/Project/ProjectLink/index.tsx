import style from './projectlink.module.css'
import Link from 'next/link'
import Image from 'next/image'
import { IProject } from '../../../types'
import { monthIndexToName } from '../../../lib/util'
import { useEffect, useRef, useState } from 'react'

const ProjectLink = ({ project }: { project: IProject }) => {
  const date =
    project.date && !project.ignoreDate ? new Date(project.date) : null
  const [imageLoaded, setImageLoaded] = useState(false)
  const mounted = useRef(false)

  useEffect(() => {
    mounted.current = true
    return () => {
      mounted.current = false
    }
  }, [])

  return (
    <div className={style['main']}>
      <Link href={'/' + project.id}>
        <a>
          <div className={style['container']}>
            <div className={style['image']}>
              <Image
                className={
                  imageLoaded
                    ? style['image']
                    : `${style['image']} ${style['loading']}`
                }
                width={48}
                height={48}
                alt={project.icon ? `${project.name} icon` : ''}
                src={
                  project.icon
                    ? '/sub/' + project.id + '/' + project.icon
                    : '/images/placeholder.png'
                }
                onLoadingComplete={() =>
                  mounted.current && setImageLoaded(true)
                }
              />
            </div>

            <div className={style['information']}>
              <h2 className={style['heading']}>{project.name}</h2>
              <p className={style['description']}>{project.description}</p>
              {date && (
                <span className={style['date']}>{`${monthIndexToName(
                  date.getMonth()
                )}, ${date.getFullYear()}`}</span>
              )}
            </div>
          </div>
        </a>
      </Link>
    </div>
  )
}

export default ProjectLink
