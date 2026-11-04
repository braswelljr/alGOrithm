/**
 * The schema.org objects the pages describe themselves with. Built from the
 * site constants and the catalog, so nothing here repeats an entry by hand.
 */
import type { Entry, LearningPath } from '@braswelljr/algorithm-visualizer/core';
import { absoluteUrl, author, repoUrl, siteDescription, siteName, siteUrl } from '../config/site.ts';

export const personLd = {
  '@type': 'Person',
  '@id': `${author.url}/#person`,
  name: author.name,
  url: author.url,
  sameAs: [author.github]
};

/** The site, its author and its source code, on every page. */
export function siteLd(): Record<string, unknown> {
  return {
    '@context': 'https://schema.org',
    '@graph': [
      {
        '@type': 'WebSite',
        '@id': `${siteUrl}/#website`,
        name: siteName,
        url: absoluteUrl('/'),
        description: siteDescription,
        inLanguage: 'en',
        author: { '@id': personLd['@id'] },
        publisher: { '@id': personLd['@id'] }
      },
      personLd,
      {
        '@type': 'SoftwareSourceCode',
        '@id': `${repoUrl}#code`,
        name: siteName,
        codeRepository: repoUrl,
        programmingLanguage: ['Go', 'TypeScript'],
        license: 'https://opensource.org/licenses/MIT',
        author: { '@id': personLd['@id'] }
      }
    ]
  };
}

/** A trail of named links, from the home page down. */
export function breadcrumbLd(trail: { name: string; path: string }[]): Record<string, unknown> {
  return {
    '@context': 'https://schema.org',
    '@type': 'BreadcrumbList',
    itemListElement: trail.map((t, i) => ({
      '@type': 'ListItem',
      position: i + 1,
      name: t.name,
      item: absoluteUrl(t.path)
    }))
  };
}

const levels: Record<string, string> = {
  beginner: 'Beginner',
  intermediate: 'Intermediate',
  advanced: 'Advanced'
};

/** One entry's documentation page, as an article that teaches its tags. */
export function entryLd(e: Entry, categoryName: string, path: string): Record<string, unknown> {
  return {
    '@context': 'https://schema.org',
    '@type': ['TechArticle', 'LearningResource'],
    headline: e.name,
    name: e.name,
    description: e.summary,
    url: absoluteUrl(path),
    inLanguage: 'en',
    about: categoryName,
    articleSection: categoryName,
    keywords: [e.name, categoryName, ...e.tags].join(', '),
    teaches: e.tags.length ? e.tags : [categoryName],
    educationalLevel: levels[e.level] ?? e.level,
    learningResourceType: 'Interactive visualization',
    isAccessibleForFree: true,
    proficiencyLevel: levels[e.level] ?? e.level,
    author: personLd,
    publisher: personLd,
    isPartOf: { '@id': `${siteUrl}/#website` }
  };
}

/** A learning path, as a free online course made of the entries it runs. */
export function pathLd(p: LearningPath, path: string): Record<string, unknown> {
  return {
    '@context': 'https://schema.org',
    '@type': 'Course',
    name: p.title,
    description: p.summary,
    url: absoluteUrl(path),
    educationalLevel: levels[p.level] ?? p.level,
    audience: { '@type': 'Audience', audienceType: p.audience },
    isAccessibleForFree: true,
    inLanguage: 'en',
    provider: personLd,
    hasCourseInstance: {
      '@type': 'CourseInstance',
      courseMode: 'online',
      courseWorkload: `${p.steps.length} steps`
    }
  };
}
