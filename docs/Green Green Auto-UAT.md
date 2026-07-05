# Green/Green Auto-UAT

Created by: Chris Marshall
Created time: May 22, 2026 6:13 AM

With AI assisted development becoming more and more pervasive, engineering velocity is increasing rapidly. New parts of the SDLC are becoming bottlenecks for the first time as other parts accelerate. In order for our organization to fully embrace this change, new paradigms must be adopted. AI has assisted in going from idea/requirements to technology design to implementation plan to Linear tickets. AI has not helped as formally with evaluation of generated implementations. This document aims to design such a system.

# Core axioms

1. Humans have not **written** the code.
2. Humans have not **read** the code.

<aside>
⚠️

This does not mean that humans CANNOT do these things, but this system does not require it.

</aside>

# Problem

# Solution overview

A separate workload which exercises the System Under Test (SUT) externally without knowledge or access to internals to validate the expected behavior of the SUT without possibility of cheating, false testing, or brittle tests of a single implementation.

## Key features

1. Natural Observation vs manipulation - Classify tests as intrusive or not. Non-intrusive tests can be run against production systems for final smoke tests while all tests can be run against lower environments.
2. Ability to use existing, running systems or start and own the systems; including dependencies and digital twins. To run tests in CI, we’ll need to start our own instance of the service. If it depends on other services, we’ll need to start those, or their [digital twins](https://factory.strongdm.ai/techniques/dtu). For smoke tests or actual UAT in deployed environments, the tests can run against an already live system.
3. Tests always describe the systems current state. This is the green/green. No concept of pending or skipped tests. If a test failure is OK, that behavior is no longer part of the SUT and delete the test. See below for a predefined behavior workflow.

# Workflows

## Predefined behavior

A hallmark of a good acceptance testing process is to define the expected behavior of a new feature before the implementation is done. Typically this is done through something like red/green TDD. Write the failing test, then make it pass. This Green/Green Auto-UAT (GGA) process doesn’t allow for failing tests. Managing pending or skipped tests is cumbersome and easier to get wrong than right. GGA relies on negative assertions to ensure the system *doesn’t* perform the expected behavior before it’s implemented.

For example, if you’re adding a new endpoint, you would assert that the new URL path **does not** return a 200 HTTP response code and that the response **does not** contain the first and last name fields.

## Two phase approach

The recommended high level flow for developing a new feature is to first fully articulate. Have healthy discussions with humans and agents to cover corner cases, etc.; include performance or other secondary effects if applicable.

Then codify these requirements as negative tests in the repository (always green).

As the feature is built, negative tests will start failing. Reverse the assertions and now that behavior is embedded into the application forever (sill green).

## Deployment

CI checks run the full suite against each PR.