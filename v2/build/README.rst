===========================
Containerized Gazette Build
===========================

See ``all.sh`` for run-able documentation of repository containers
and build processes.

Build and Publish the Base Runtime Image
========================================

The ``liveramp/gazette-base`` image contains required Gazette build & runtime
dependencies, notably compiled RocksDB libraries and a protobuf compiler. This
image is used both for building gazette itself, and also for multi-stage
Dockerfiles which pick out compiled binaries of the Gazette build but still
require the RocksDB runtime.

Typically developers should use the published ``liveramp/gazette-base`` image,
but may build their own image via:

.. code-block:: console

    $ docker build . -f build/Dockerfile.gazette-base --tag liveramp/gazette-base

Test it locally by temporarily changing the references in
``build/Dockerfile.gazette-build`` and ``build/cmd/Dockerfile.gazette`` from
``liveramp/gazette-base:X.Y.Z`` [*]_ to ``liveramp/gazette-base:latest``. Then
run the ``docker build`` commands given above.

Once the image has been tested and is ready to publish, pick an appropriate
`semantic version number`_ [*]_ then tag and push the image:

.. code-block:: console

    $ docker tag liveramp/gazette-base liveramp/gazette-base:X.Y.Z
    $ docker push liveramp/gazette-base:X.Y.Z

It may be necessary to log in to Docker Hub before ``docker push``:

.. code-block:: console

    $ docker login
    # Interactively enter username and password.

.. _semantic version number: https://semver.org

.. [*] Note that this project's convention is to not prefix the version number
       with "v".
.. [*] For example, changing major versions of Go would be a major version
       bump, installing an additional tool would be a minor version bump, and
       fixing a bug in a ``RUN`` command would be a patch version bump.
