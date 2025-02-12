/**
 * Creating a sidebar enables you to:
 - create an ordered group of docs
 - render a sidebar for each doc of that group
 - provide next/previous navigation

 The sidebars can be generated from the filesystem, or explicitly defined here.

 Create as many sidebars as you want.
 */

// @ts-check

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  // By default, Docusaurus generates a sidebar from the docs folder structure
  // tutorialSidebar: [
  //   {type: 'autogenerated', dirName: '.'},
  // ],

  // But you can create a sidebar manually
  tutorialSidebar: [
    {
      type: 'category',
      label: 'About',
      items: [
        'about/about',
        // 'about/comparison',
        // 'about/implementation'
      ],
    },
    {
      label: 'Concepts',
      id: 'concepts/concepts',
      type: 'doc',
    },
    {
      type: 'category',
      label: 'Installation',
      items: [
        'installation/prerequisites',
        'installation/install',
        'installation/configuration',
        {
          type: 'category',
          label: 'Demo',
          items: [
            'installation/demo/new_module',
            'installation/demo/module',
            'installation/demo/edit_module',
            'installation/demo/logs',
            'installation/demo/feedback'
          ],
        }
      ],
    },
    {
      type: 'category',
      label: 'Writing Templates',
      items: [
        'templates/templates',
        'templates/validations',
        'templates/dependencies',
      ],
    },
    {
      label: 'Roadmap',
      id: 'roadmap/roadmap',
      type: 'doc',
    },
  ],
};

module.exports = sidebars;
