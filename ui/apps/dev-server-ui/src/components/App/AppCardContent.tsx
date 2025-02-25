import { Link } from '@inngest/components/Link/Link';
import { type AppKind } from '@inngest/components/types/app';
import { RiExternalLinkLine } from '@remixicon/react';

import { type GetAppsQuery } from '@/store/generated';
import UpdateApp from './UpdateApp';

const getAppCardContent = ({ app }: { app: GetAppsQuery['apps'][number] }) => {
  const appKind: AppKind = !app.connected ? 'error' : app.functionCount > 0 ? 'primary' : 'warning';

  const status = !app.connected
    ? 'Not Synced'
    : app.functionCount === 0
    ? 'No functions found'
    : null;

  const footerHeader = !app.connected ? (
    app.error === 'unreachable' ? (
      `The Inngest Dev Server can't find your application.`
    ) : (
      `Error: ${app.error}`
    )
  ) : app.functionCount === 0 ? (
    'There are currently no functions registered at this URL.'
  ) : (
    <div className="flex w-full items-center justify-between">
      <span>
        {app.functionCount} {app.functionCount === 1 ? 'function' : 'functions'} found
      </span>
      <Link size="small" href="/functions" arrowOnHover>
        View functions
      </Link>
    </div>
  );

  const footerContent = !app.connected ? (
    <>
      <p className="text-subtle pb-4">
        Ensure your full URL is correct, including the port. Inngest automatically scans{' '}
        <span className="text-basis">multiple ports</span> by default.
      </p>
      <UpdateApp app={app} />
    </>
  ) : app.functionCount === 0 ? (
    <>
      <p className="text-subtle pb-4">
        Ensure you have created a function and are exporting it correctly from your serve() command.
      </p>
      <Link
        size="small"
        target="_blank"
        href="https://www.inngest.com/docs/learn/serving-inngest-functions?ref=dev-app"
        iconAfter={<RiExternalLinkLine className="h-4 w-4" />}
      >
        How to serve functions
      </Link>
    </>
  ) : (
    <ul className="divide-subtle divide-y">
      {[...app.functions]
        .sort((a, b) => a.name.localeCompare(b.name))
        .map((func) => {
          return (
            <li key={func.id} className="text-subtle py-2">
              {func.name}
            </li>
          );
        })}
    </ul>
  );

  return { appKind, status, footerHeader, footerContent };
};

export default getAppCardContent;
