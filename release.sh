#!/bin/bash

if [ $# -lt 1 ]; then
    echo "Please provide a semantic version, 'major', 'minor' or 'patch'"
    exit 1
fi


NEW_VERSION="$1"
TAG_COMMIT='HEAD'

# Fetch existing tags before choosing what the next one should be.
echo 'Fetching existing tags...'
git fetch --tags

# Parse major, minor and patch groups of latest tag.
LATEST_RELEASE_TAG=$(git tag | grep -iE 'v.*' | sort -V | tail -n 1)
if [[ -z "${LATEST_RELEASE_TAG}" ]]; then
    echo "No release tags found. Please create an initial release tag first."
    exit 1
fi

NEW_RELEASE_MAJOR=$(echo ${LATEST_RELEASE_TAG//v} | cut -d. -f 1)
NEW_RELEASE_MINOR=$(echo ${LATEST_RELEASE_TAG//v} | cut -d. -f 2)
NEW_RELEASE_PATCH=$(echo ${LATEST_RELEASE_TAG//v} | cut -d. -f 3)

# Set the major, minor and patch groups of the new tag.
if [[ "${NEW_VERSION}" == 'major' ]]; then
    NEW_RELEASE_MAJOR=$(($NEW_RELEASE_MAJOR + 1))
    NEW_RELEASE_MINOR='0'
    NEW_RELEASE_PATCH='0'
elif [[ "${NEW_VERSION}" == 'minor' ]]; then
    NEW_RELEASE_MINOR=$(($NEW_RELEASE_MINOR + 1))
    NEW_RELEASE_PATCH='0'
elif [[ "${NEW_VERSION}" == 'patch' ]]; then
    NEW_RELEASE_PATCH=$(($NEW_RELEASE_PATCH + 1))
elif [[ "${NEW_VERSION}" =~ v.*\..*\..* ]]; then
    NEW_RELEASE_MAJOR=$(echo ${NEW_VERSION//v} | cut -d. -f 1)
    NEW_RELEASE_MINOR=$(echo ${NEW_VERSION//v} | cut -d. -f 2)
    NEW_RELEASE_PATCH=$(echo ${NEW_VERSION//v} | cut -d. -f 3)
else
    echo "Invalid argument. Please provide a semantic version, 'major', 'minor' or 'patch'"
    exit 1
fi

# Construct the new release tag.
echo "Latest release tag is '${LATEST_RELEASE_TAG}'"
NEW_RELEASE="v${NEW_RELEASE_MAJOR}.${NEW_RELEASE_MINOR}.${NEW_RELEASE_PATCH}"
echo "Creating release '${NEW_RELEASE}'"

# Construct the tag message.
# If the tag message is empty, use the default message.
TAG_MESSAGE_FILE=$(mktemp)
TAG_MESSAGE_HEADER="Version ${NEW_RELEASE_MAJOR}.${NEW_RELEASE_MINOR}.${NEW_RELEASE_PATCH}"
TAG_MESSAGE_CHANGES=$(git log --pretty=format:"- %s" -n 20 "${LATEST_RELEASE_TAG}..${TAG_COMMIT}")

echo "${TAG_MESSAGE_HEADER}" > "${TAG_MESSAGE_FILE}"
echo '' >> "${TAG_MESSAGE_FILE}"
echo "Commits since ${LATEST_RELEASE_TAG}:" >> "${TAG_MESSAGE_FILE}"
echo "${TAG_MESSAGE_CHANGES}" >> "${TAG_MESSAGE_FILE}"
${EDITOR} "${TAG_MESSAGE_FILE}"

git tag -a "${NEW_RELEASE}" -F "${TAG_MESSAGE_FILE}" "${TAG_COMMIT}"

while true; do
    read -p "Release tag created, push, delete or keep? [${bold}P${normal}ush/${bold}K${normal}eep/${bold}D${normal}elete]: " choice
    case $choice in
        [Pp]*) git push origin ${NEW_RELEASE} ; break ;;
        [Kk]*) break ;;
        [Dd]*) git tag -d ${NEW_RELEASE} ; break ;;
    esac
done